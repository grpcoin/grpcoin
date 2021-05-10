// Copyright 2021 Ahmet Alp Balkan
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
