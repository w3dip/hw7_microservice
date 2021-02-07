package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

type BizServerImpl struct {
	mu  sync.RWMutex
	ctx context.Context
	acl string
}

func NewBizManager(ctx context.Context, acl string) *BizServerImpl {
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

func StartMyMicroservice(ctx context.Context, addr string, acl string) error {

	//wg := &sync.WaitGroup{}
	//wg.Add(1)
	fmt.Println("Run server")
	go func( /*wg *sync.WaitGroup, */ ctx context.Context, addr string /*server *grpc.Server, lis net.Listener*/) {
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalln("cant listet port", err)
		}

		server := grpc.NewServer()

		RegisterBizServer(server, NewBizManager(ctx, acl))
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

	}( /*wg, */ ctx /*server, lis*/, addr)

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
