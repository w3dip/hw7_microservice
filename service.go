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
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("cant listet port", err)
	}

	server := grpc.NewServer()

	RegisterBizServer(server, NewBizManager(ctx, acl))

	fmt.Printf("starting server at %s", addr)
	err = server.Serve(lis)
	if err != nil {
		return err
	}
	return nil
}
