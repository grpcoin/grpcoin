package main

import (
	"context"
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
	pb.RegisterAccountServer(srv, newAccountService(accountServiceOpts{}))
	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()
	err = srv.Serve(lis)
	panic(err)
}
