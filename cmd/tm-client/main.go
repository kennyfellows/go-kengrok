package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "go-kengrok/proto/tunnel-manager"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
  conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

  if err != nil {
    fmt.Println( "Error connecting to grpc server ", err.Error() )
  }

  defer conn.Close()

	c := pb.NewTunnelManagerClient(conn)

  ctx, cancel := context.WithTimeout( context.Background(), time.Second )

	defer cancel()

	r, err := c.CreateTunnel( ctx, &pb.CreateTunnelRequest{Subdomain: "foobars"} )

  if err != nil {
		log.Fatalf("could not create tunnel: %v", err)
	}

  log.Printf("Tunnel created: %v", r.GetPort() )
}
