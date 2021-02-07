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
)

//type ACL struct {
//	data map[string][]string
//}

type ServerImpl struct {
	mu     sync.RWMutex
	ctx    context.Context
	acl    map[string][]string
	method string
	host   string
}

func NewBizManager(ctx context.Context, acl map[string][]string) *ServerImpl {
	return &ServerImpl{
		mu:  sync.RWMutex{},
		ctx: ctx,
		acl: acl,
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

func NewAdminManager(ctx context.Context, acl map[string][]string) *ServerImpl {
	return &ServerImpl{
		mu:  sync.RWMutex{},
		ctx: ctx,
		acl: acl,
	}
}

func (server *ServerImpl) Logging(nothing *Nothing, loggingServer Admin_LoggingServer) error {
	ctx := loggingServer.Context()
	md, _ := metadata.FromIncomingContext(ctx)
	consumer := md.Get("consumer")
	if consumer == nil || len(consumer) == 0 {
		return status.Errorf(codes.DataLoss, "can't find consumer in request")
	}
	//server.mu.Lock()
	//defer server.mu.Unlock()
	err := loggingServer.Send(&Event{Consumer: consumer[0], Method: server.method, Host: server.host})
	if err != nil {
		return err
	}
	return nil
}

func (server *ServerImpl) Statistics(statInterval *StatInterval, statisticsServer Admin_StatisticsServer) error {
	return nil
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

		RegisterBizServer(server, NewBizManager(ctx, acl))
		RegisterAdminServer(server, NewAdminManager(ctx, acl))
		//defer wg.Done()

		//fmt.Printf("starting server at %s", addr)
		//fmt.Printf("Try to start")
		go server.Serve(lis)
		//if err != nil {
		//	log.Fatalln("cant start server", err)
		//}
		fmt.Println("Server is running")
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
		server.mu.Lock()
		defer server.mu.Unlock()
		acl := server.acl
		user := consumer[0]
		fmt.Printf("Checking permissions for user %v\n", user)
		if val, ok := acl[user]; ok {
			fmt.Printf("Found rules %v\n", val)
			fmt.Printf("Request method %v\n", info.FullMethod)
			for _, expr := range val {
				if ok, _ := filepath.Match(expr, info.FullMethod); ok {
					fmt.Printf("Rules for user %v found successfully\n", user)
					server.method = info.FullMethod
					p, _ := peer.FromContext(ctx)
					server.host = p.Addr.String()
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
		server.mu.Lock()
		defer server.mu.Unlock()
		acl := server.acl
		user := consumer[0]
		fmt.Printf("Checking permissions for user %v\n", user)
		if val, ok := acl[user]; ok {
			fmt.Printf("Found rules %v\n", val)
			fmt.Printf("Request method %v\n", info.FullMethod)
			for _, expr := range val {
				if ok, _ := filepath.Match(expr, info.FullMethod); ok {
					fmt.Printf("Rules for user %v found successfully\n", user)
					server.method = info.FullMethod
					p, _ := peer.FromContext(ctx)
					server.host = p.Addr.String()
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
