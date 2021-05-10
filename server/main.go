package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/ahmetb/grpcoin/api/grpcoin"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := os.Getenv("LISTEN_ADDR")

	lis, err := net.Listen("tcp", addr+":"+port)
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	as, err := newAccountService(accountServiceOpts{
		redisIP: os.Getenv("REDIS_IP"),
	})
	if err != nil {
		panic(err)
	}
	pb.RegisterAccountServer(srv, as)
	pb.RegisterTickerInfoServer(srv, new(tickerService))
	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()
	err = srv.Serve(lis)
	if err != nil {
		panic(err)
	}
	fmt.Println("gracefully shut down the server")
}
