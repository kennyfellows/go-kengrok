package main

import (
	"context"
	"fmt"
	pb "go-kengrok/proto/proxy-manager"
	"log"
	"net"

	"go-kengrok/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
  "google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
  pb.UnimplementedProxyManagerServer
}

func subdomainInUse( subdomain string ) bool {
  redisClient := utils.GetRedisClient()
  ctx := context.Background()
  key := fmt.Sprintf( "kengrok-map:%s", subdomain )
  exists, err := redisClient.Exists( ctx, key ).Result()

  if err != nil {
    log.Fatal("Error checking if subdomain is already in use")
  }

  return exists == 1
}

func savePortMapping( subdomain string, port int32 ) ( ok bool, err error ) {
  redisClient := utils.GetRedisClient()
  ctx := context.Background()
  key := fmt.Sprintf( "kengrok-map:%s", subdomain )
  redisClient.Set( ctx, key, port, 0 ).Err()

  return true, nil
}

func (s *server) StartProxy (ctx context.Context, req *pb.StartProxyRequest)(*pb.StartProxyResponse, error) {
  subdomain := req.GetSubdomain()
  port := req.GetPort()

  subdomainIsInUse := subdomainInUse( subdomain )

  if subdomainIsInUse {
    return nil, status.Error( codes.InvalidArgument, "Subdomain is already in use" )
  } else {
    savePortMapping( subdomain, port )
  }

  log.Printf("Received mapping request for: %v -> %v", subdomain, port )
  return &pb.StartProxyResponse{ Port: 4567 }, nil
}

func (s *server) EndProxy (ctx context.Context, req *pb.EndProxyRequest)(*emptypb.Empty, error) {
  subdomain := req.GetSubdomain()
  redisClient := utils.GetRedisClient()
	context := context.Background()
  key := fmt.Sprintf( "kengrok-map:%s", subdomain )

  _, err := redisClient.Del( context, key ).Result()

  if err != nil {
    return nil, status.Error( codes.Internal, "Error removing mapping" )
  }

  return &emptypb.Empty{}, nil
}

func main() {
  lis, err := net.Listen( "tcp", ":50051" )

  if err != nil {
    log.Fatalf( "Failed to listen: %v", err )
  }

  s := grpc.NewServer()
  pb.RegisterProxyManagerServer( s, &server{} )
  log.Printf( "server listening at %v", lis.Addr() )

  if err := s.Serve( lis ); err != nil {
    log.Fatalf( "Failed to serve: %v", err )
  }
}
