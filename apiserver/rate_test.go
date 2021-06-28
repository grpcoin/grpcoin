package main

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func Test_findIP(t *testing.T) {
	tests := []struct {
		name    string
		rpcCtx  context.Context
		want    string
		wantErr bool
	}{
		{name: "empty ctx",
			rpcCtx:  context.TODO(),
			want:    "",
			wantErr: false},
		{name: "xri hdr",
			rpcCtx: metadata.NewIncomingContext(context.Background(),
				metadata.New(map[string]string{"x-real-ip": "1.2.3.4"})),
			want:    "1.2.3.4",
			wantErr: false},
		{name: "xff hdr single ip",
			rpcCtx: metadata.NewIncomingContext(context.Background(),
				metadata.New(map[string]string{"x-forwarded-for": "1.2.3.4"})),
			want:    "1.2.3.4",
			wantErr: false},
		{name: "xff hdr multiple ips",
			rpcCtx: metadata.NewIncomingContext(context.Background(),
				metadata.New(map[string]string{"x-forwarded-for": "1.2.3.4, 5.6.7.8, y"})),
			want:    "1.2.3.4",
			wantErr: false},
		{name: "xri over xff",
			rpcCtx: metadata.NewIncomingContext(context.Background(),
				metadata.New(map[string]string{
					"x-real-ip":       "0.0.0.0",
					"x-forwarded-for": "1.2.3.4, 5.6.7.8, y"})),
			want:    "0.0.0.0",
			wantErr: false},
		{name: "from peer info",
			rpcCtx: peer.NewContext(context.Background(),
				&peer.Peer{Addr: &net.TCPAddr{IP: net.IPv4(1, 1, 1, 1), Port: 54321}}),
			want:    "1.1.1.1",
			wantErr: false},
		{name: "from peer info",
			rpcCtx: peer.NewContext(context.Background(),
				&peer.Peer{Addr: fakeNetAddr("hello")}),
			wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findIP(tt.rpcCtx)
			if (err != nil) != tt.wantErr {
				t.Errorf("findIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findIP() got = %v, want %v", got, tt.want)
			}
		})
	}
}

type fakeNetAddr string
func (f fakeNetAddr) Network() string {return "fake"}
func (f fakeNetAddr) String() string {return string(f)}

