package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

//type ACL struct {
//	data map[string][]string
//}

type BizServerImpl struct {
	mu  sync.RWMutex
	ctx context.Context
	acl map[string][]string
}

func NewBizManager(ctx context.Context, acl map[string][]string) *BizServerImpl {
	return &BizServerImpl{
		mu:  sync.RWMutex{},
		ctx: ctx,
		acl: acl,
	}
}

func (server *BizServerImpl) Check(ctx context.Context, nothing *Nothing) (*Nothing, error) {
	return nil, nil
}

func (server *BizServerImpl) Add(ctx context.Context, nothing *Nothing) (*Nothing, error) {
	return nil, nil
}

func (server *BizServerImpl) Test(ctx context.Context, nothing *Nothing) (*Nothing, error) {
	return nil, nil
}

func (server *BizServerImpl) mustEmbedUnimplementedBizServer() {
}

type AdminServerImpl struct {
	mu  sync.RWMutex
	ctx context.Context
	acl map[string][]string
}

func NewAdminManager(ctx context.Context, acl map[string][]string) *AdminServerImpl {
	return &AdminServerImpl{
		mu:  sync.RWMutex{},
		ctx: ctx,
		acl: acl,
	}
}

func (server *AdminServerImpl) Logging(nothing *Nothing, loggingServer Admin_LoggingServer) error {
	return nil
}

func (server *AdminServerImpl) Statistics(statInterval *StatInterval, statisticsServer Admin_StatisticsServer) error {
	return nil
}

func (server *AdminServerImpl) mustEmbedUnimplementedAdminServer() {
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
			log.Fatalln("cant listet port", err)
		}

		server := grpc.NewServer()

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
