package clients

import (
	"context"
	"fmt"

	pb "go-kengrok/proto/proxy-manager"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var instance ProxyClient

type ProxyClient interface {
	StartProxyMapping(ctx context.Context, subdomain string, port int) error
	EndProxyMapping( ctx context.Context, subdomain string ) error
}

type grpcProxyClient struct {
	conn   *grpc.ClientConn
	client pb.ProxyManagerClient
}

func (proxyClient *grpcProxyClient) StartProxyMapping(ctx context.Context, subdomain string, portInt int) error {
	port := int32(portInt)
	request := &pb.StartProxyRequest{
		Port:      port,
		Subdomain: subdomain,
	}

  _, err := proxyClient.client.StartProxy(ctx, request)

	return err
}

func (proxyClient *grpcProxyClient) EndProxyMapping(ctx context.Context, subdomain string) error {
	request := &pb.EndProxyRequest{
		Subdomain: subdomain,
	}

  _, err := proxyClient.client.EndProxy(ctx, request)

	return err
}

func newClient() (ProxyClient, error) {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, fmt.Errorf("Failed to connect to gRPC server: %v", err)
	}

	return &grpcProxyClient{
		conn:   conn,
		client: pb.NewProxyManagerClient(conn),
	}, nil
}

func GetProxyManagerClient() (ProxyClient, error) {

	if instance != nil {
		return instance, nil
	}

	var err error
	instance, err = newClient()

	return instance, err
}
