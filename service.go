package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"path/filepath"
	"sync"
	"time"
)

//type ACL struct {
//	data map[string][]string
//}

type ServerImpl struct {
	mu  sync.RWMutex
	ctx context.Context
	acl map[string][]string

	loggers      []chan *Event
	createLogger chan chan *Event
	streamLogger chan *Event

	statistics      []chan *Stat
	createStatistic chan chan *Stat
	streamStatistic chan *Stat
	//method string
	//host   string
	//stat   *Stat
}

func NewManager(ctx context.Context, acl map[string][]string) *ServerImpl {
	return &ServerImpl{
		mu:              sync.RWMutex{},
		ctx:             ctx,
		acl:             acl,
		streamLogger:    make(chan *Event, 0),
		createLogger:    make(chan chan *Event, 0),
		streamStatistic: make(chan *Stat, 0),
		createStatistic: make(chan chan *Stat, 0),
	}
}

func (server *ServerImpl) Check(ctx context.Context, nothing *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (server *ServerImpl) Add(ctx context.Context, nothing *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (server *ServerImpl) Test(ctx context.Context, nothing *Nothing) (*Nothing, error) {
	return &Nothing{}, nil
}

func (server *ServerImpl) mustEmbedUnimplementedBizServer() {
}

//type AdminServerImpl struct {
//	mu  sync.RWMutex
//	ctx context.Context
//	acl map[string][]string
//}

//func NewAdminManager(ctx context.Context, acl map[string][]string) *ServerImpl {
//	return &ServerImpl{
//		mu:  sync.RWMutex{},
//		ctx: ctx,
//		acl: acl,
//	}
//}

func (server *ServerImpl) Logging(nothing *Nothing, loggingServer Admin_LoggingServer) error {
	//ctx := loggingServer.Context()
	//md, _ := metadata.FromIncomingContext(ctx)
	//consumer := md.Get("consumer")
	//if consumer == nil || len(consumer) == 0 {
	//	return status.Errorf(codes.DataLoss, "can't find consumer in request")
	//}
	////server.mu.Lock()
	////defer server.mu.Unlock()
	//err := loggingServer.Send(&Event{Consumer: consumer[0], Method: server.method, Host: server.host})
	//if err != nil {
	//	return err
	//}
	ch := make(chan *Event, 0)
	server.createLogger <- ch

	for {
		select {
		case event := <-ch:
			loggingServer.Send(event)
		case <-server.ctx.Done():
			return nil
		}
	}
}

func (server *ServerImpl) Statistics(statInterval *StatInterval, statisticsServer Admin_StatisticsServer) error {
	//ctx := statisticsServer.Context()
	//go func(ctx context.Context, statInterval *StatInterval, statisticsServer Admin_StatisticsServer, ) {
	//
	//}(ctx, statInterval, statisticsServer)
	//interval := int64(statInterval.IntervalSeconds) * time.Second.Milliseconds()
	//ticker := time.Tick(2 * time.Second)
	//for _ = range ticker {
	//	//server.mu.Lock()
	//	//fmt.Printf("Current stat %v\n", server.stat)
	//	//err := statisticsServer.Send(server.stat)
	//	stat := &Stat{
	//		ByMethod: map[string]uint64{
	//			"/main.Biz/Check":        1,
	//			"/main.Biz/Add":          1,
	//			"/main.Biz/Test":         1,
	//			"/main.Admin/Statistics": 1,
	//		},
	//		ByConsumer: map[string]uint64{
	//			"biz_user":  2,
	//			"biz_admin": 1,
	//			"stat":      1,
	//		},
	//	}
	//	err := statisticsServer.Send(stat)
	//	if err != nil {
	//		return err
	//	}
	//	//server.mu.Unlock()
	//}
	ch := make(chan *Stat, 0)
	server.createStatistic <- ch

	ticker := time.NewTicker(time.Second * time.Duration(statInterval.IntervalSeconds))
	sum := &Stat{
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
	}
	for {
		select {
		case <-ticker.C:
			statisticsServer.Send(sum)
			sum = &Stat{
				ByMethod:   make(map[string]uint64),
				ByConsumer: make(map[string]uint64),
			}
		case stat := <-ch:
			for k, v := range stat.ByMethod {
				sum.ByMethod[k] += v
			}
			for k, v := range stat.ByConsumer {
				sum.ByConsumer[k] += v
			}

		case <-server.ctx.Done():
			return nil
		}
	}
	//return nil
}

func (server *ServerImpl) mustEmbedUnimplementedAdminServer() {
}

func StartMyMicroservice(ctx context.Context, addr string, acl string) error {
	parsedACL := make(map[string][]string)
	err := json.Unmarshal([]byte(acl), &parsedACL)
	if err != nil {
		//log.Fatalln("cant parse acl", err)
		return err
	}
	//wg := &sync.WaitGroup{}
	//wg.Add(1)
	fmt.Println("Run server")
	go func( /*wg *sync.WaitGroup, */ ctx context.Context, addr string, acl map[string][]string /*server *grpc.Server, lis net.Listener*/) {
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalln("cant listen port", err)
		}

		server := grpc.NewServer(
			grpc.UnaryInterceptor(authInterceptor),
			grpc.StreamInterceptor(authStreamInterceptor),
		)

		manager := NewManager(ctx, acl)
		RegisterBizServer(server, manager)
		RegisterAdminServer(server, manager)
		//defer wg.Done()

		//fmt.Printf("starting server at %s", addr)
		//fmt.Printf("Try to start")
		go server.Serve(lis)
		//if err != nil {
		//	log.Fatalln("cant start server", err)
		//}
		fmt.Println("Server is running")

		go func() {
			for {
				select {
				case ch := <-manager.createLogger:
					manager.loggers = append(manager.loggers, ch)
				case event := <-manager.streamLogger:
					for _, ch := range manager.loggers {
						ch <- event
					}
				case <-ctx.Done():
					return
				}
			}
		}()

		go func() {
			for {
				select {
				case ch := <-manager.createStatistic:
					manager.statistics = append(manager.statistics, ch)
				case stat := <-manager.streamStatistic:
					for _, ch := range manager.statistics {
						ch <- stat
					}
				case <-ctx.Done():
					return
				}
			}
		}()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("finish called")
				server.Stop()
				//time.Sleep(time.Duration(2) * 10 * time.Millisecond)
				//lis.Close()
				return
				//default:
				//	fmt.Printf("noop")
			}
		}

	}( /*wg, */ ctx /*server, lis*/, addr, parsedACL)

	//time.Sleep(time.Duration(2) * 1000 * time.Millisecond)

	//select {
	//		case <-ctx.Done():
	//			fmt.Printf("finish called")
	//			server. GracefulStop()
	//			lis.Close()
	//			return nil
	//			//default:
	//		}
	//<-ctx.Done()

	//server.Stop()
	//lis.Close()

	//for {
	//	select {
	//	case <-ctx.Done():
	//		fmt.Printf("finish called")
	//		server. /*Graceful*/ Stop()
	//		return nil
	//		//default:
	//	}
	//}
	//fmt.Printf("Waiting for ending...")
	fmt.Println("Waiting for termination")
	//wg.Wait()
	return nil
}

func authStreamInterceptor(srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler) error {
	ctx := ss.Context()

	md, _ := metadata.FromIncomingContext(ctx)
	consumer := md.Get("consumer")
	if consumer == nil {
		return status.Error(codes.Unauthenticated, "Auth data is empty")
	}

	server, ok := srv.(*ServerImpl)
	if ok {
		//switch serverType := info.Server.(type) {
		//case BizServerImpl:
		//server := info.Server.(BizServerImpl)
		//server.mu.Lock()
		//defer server.mu.Unlock()
		acl := server.acl
		user := consumer[0]
		method := info.FullMethod
		fmt.Printf("Checking permissions for user %v\n", user)
		if val, ok := acl[user]; ok {
			fmt.Printf("Found rules %v\n", val)
			fmt.Printf("Request method %v\n", method)
			for _, expr := range val {
				if ok, _ := filepath.Match(expr, method); ok {
					fmt.Printf("Rules for user %v found successfully\n", user)
					p, _ := peer.FromContext(ctx)
					host := p.Addr.String()
					server.streamStatistic <- &Stat{
						ByConsumer: map[string]uint64{user: 1},
						ByMethod:   map[string]uint64{method: 1},
					}

					server.streamLogger <- &Event{
						Consumer: user,
						Method:   method,
						Host:     host,
					}
					err := handler(srv, ss)
					if err != nil {
						return err
					}
					return nil
				}
			}
		}
		fmt.Printf("Rules for user %v not found\n", user)
		//default:
		//	fmt.Printf("server type is %v", serverType)
	}
	return status.Error(codes.Unauthenticated, "User unknown")
}

func authInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	md, _ := metadata.FromIncomingContext(ctx)
	consumer := md.Get("consumer")
	if consumer == nil {
		return nil, status.Error(codes.Unauthenticated, "Auth data is empty")
	}
	reply, err := handler(ctx, req)
	fmt.Printf(`--
	after incoming call=%v req=%#v
	reply=%#v
	md=%v
	err=%v

`, info.FullMethod, req, reply, md, err)
	server, ok := info.Server.(*ServerImpl)
	if ok {
		//switch serverType := info.Server.(type) {
		//case BizServerImpl:
		//server := info.Server.(BizServerImpl)
		//server.mu.Lock()
		//defer server.mu.Unlock()
		acl := server.acl
		user := consumer[0]
		method := info.FullMethod
		fmt.Printf("Checking permissions for user %v\n", user)
		if val, ok := acl[user]; ok {
			fmt.Printf("Found rules %v\n", val)
			fmt.Printf("Request method %v\n", method)
			for _, expr := range val {
				if ok, _ := filepath.Match(expr, method); ok {
					fmt.Printf("Rules for user %v found successfully\n", user)
					p, _ := peer.FromContext(ctx)
					host := p.Addr.String()

					//currentVal := server.stat.ByMethod[info.FullMethod]
					//server.stat.ByMethod[info.FullMethod] = currentVal + 1
					server.streamStatistic <- &Stat{
						ByConsumer: map[string]uint64{user: 1},
						ByMethod:   map[string]uint64{method: 1},
					}

					server.streamLogger <- &Event{
						Consumer: user,
						Method:   method,
						Host:     host,
					}
					return reply, err
				}
			}
		}
		fmt.Printf("Rules for user %v not found\n", user)
		//default:
		//	fmt.Printf("server type is %v", serverType)
	}
	return nil, status.Error(codes.Unauthenticated, "User unknown")
	//return reply, err
}
