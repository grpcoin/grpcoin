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
	"crypto/tls"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ahmetb/grpcoin/api/grpcoin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func main() {
	log.SetFlags(log.Lmicroseconds | log.Ltime)
	url := `grpcoin-main-kafjc7sboa-wl.a.run.app:443`
	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	creds := credentials.NewTLS(&tls.Config{})
	conn, err := grpc.DialContext(ctx, url, grpc.WithTransportCredentials(creds))
	if err != nil {
		panic(err)
	}
	client := grpcoin.NewTickerInfoClient(conn)
	for ctx.Err() == nil {
		log.Printf("connecting")
		stream, err := client.Watch(ctx, &grpcoin.Ticker{Ticker: "BTC-USD"})
		if err != nil {
			panic(err)
		}
		for ctx.Err() == nil {
			msg, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				if v, ok := status.FromError(err); ok && v.Code() == codes.Canceled {
					break
				}
				log.Fatalf("unexpected: %v", err)
			}
			log.Printf("[server:%s] --  %d.%d",
				msg.T.AsTime().Format(time.RFC3339Nano),
				msg.Price.GetUnits(),
				msg.Price.GetNanos())
		}
		log.Printf("disconnected")
		time.Sleep(time.Second)
	}
}
