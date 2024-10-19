package main

import (
	"context"
	pb "go-kengrok/proto/tunnel-manager"
	"log"
	"net"

	"google.golang.org/grpc"
)

type server struct {
  pb.UnimplementedTunnelManagerServer
}

func (s *server) RequestTunnel (ctx context.Context, req *pb.RequestTunnelRequest)(*pb.RequestTunnelResponse, error) {
  log.Printf("Received: %v", req.GetSubdomain() )

  return &pb.RequestTunnelResponse{ Port: 4567 }, nil
}

func main() {
  lis, err := net.Listen( "tcp", ":50051" )

  if err != nil {
    log.Fatalf( "Failed to listen: %v", err )
  }

  s := grpc.NewServer()
  pb.RegisterTunnelManagerServer( s, &server{} )
  log.Printf( "server listening at %v", lis.Addr() )

  if err := s.Serve( lis ); err != nil {
    log.Fatalf( "Failed to serve: %v", err )
  }
}