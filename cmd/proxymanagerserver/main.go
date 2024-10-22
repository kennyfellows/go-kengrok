package main

import (
	"context"
	"fmt"
	pb "github.com/kennyfellows/go-kengrok/proto/proxy-manager"
	"log"
	"net"
  "time"

  "github.com/redis/go-redis/v9"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
  "google.golang.org/protobuf/types/known/emptypb"
)

type server struct {
  pb.UnimplementedProxyManagerServer
  redisClient *redis.Client
}

func getRedisClient() ( *redis.Client, error ) {
  redisClient := redis.NewClient( &redis.Options{
    Addr: "localhost:6379",
  })

  ctx, cancel := context.WithTimeout( context.Background(), 5*time.Second )

  defer cancel()

  _, err := redisClient.Ping( ctx ).Result()

  return redisClient, err
}

func subdomainInUse( r *redis.Client, subdomain string ) bool {
  ctx := context.Background()
  key := fmt.Sprintf( "kengrok-map:%s", subdomain )
  exists, err := r.Exists( ctx, key ).Result()

  if err != nil {
    log.Fatal("Error checking if subdomain is already in use")
  }

  return exists == 1
}

func savePortMapping( r *redis.Client, subdomain string, port int32 ) ( ok bool, err error ) {
  ctx := context.Background()
  key := fmt.Sprintf( "kengrok-map:%s", subdomain )
  r.Set( ctx, key, port, 0 ).Err()

  return true, nil
}

func (s *server) StartProxy (ctx context.Context, req *pb.StartProxyRequest)(*pb.StartProxyResponse, error) {
  subdomain := req.GetSubdomain()
  port := req.GetPort()

  subdomainIsInUse := subdomainInUse( s.redisClient, subdomain )

  if subdomainIsInUse {
    return nil, status.Error( codes.InvalidArgument, "Subdomain is already in use" )
  } else {
    savePortMapping( s.redisClient, subdomain, port )
  }

  log.Printf("Received mapping request for: %v -> %v", subdomain, port )

  //TODO: remove the port number from response, it is unnecessary
  return &pb.StartProxyResponse{ Port: 4567 }, nil
}

func (s *server) EndProxy (ctx context.Context, req *pb.EndProxyRequest)(*emptypb.Empty, error) {
  subdomain := req.GetSubdomain()
	context := context.Background()
  key := fmt.Sprintf( "kengrok-map:%s", subdomain )

  _, err := s.redisClient.Del( context, key ).Result()

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
  r, err := getRedisClient()

  if err != nil {
    log.Fatalf("Error connecting to redis: %v", err)
  }

  pb.RegisterProxyManagerServer( s, &server{
    redisClient: r,
  })

  log.Printf( "server listening at %v", lis.Addr() )

  if err := s.Serve( lis ); err != nil {
    log.Fatalf( "Failed to serve: %v", err )
  }
}
