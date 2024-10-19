package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
  "time"
  "strconv"

  "google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc"
	pb "go-kengrok/proto/tunnel-manager"
	"golang.org/x/crypto/ssh"
)

type ReverseTunnel struct {
	RemoteHost string
	RemotePort int
	LocalHost  string
	SSHConfig  *ssh.ClientConfig
}

func (tun *ReverseTunnel) Start( subdomain string ) (int, error) {

	connString := fmt.Sprintf("%s:%d", tun.RemoteHost, tun.RemotePort)
	sshConn, err := ssh.Dial("tcp", connString, tun.SSHConfig)

	if err != nil {
		return 0, fmt.Errorf("Failed to connect to SSH server: %v", err)
	}

	defer sshConn.Close()

	listener, err := sshConn.Listen("tcp", "0.0.0.0:0")

	if err != nil {
		return 0, fmt.Errorf("Failed to request remote port forward: %v", err)
	}

	defer listener.Close()

	_, portStr, _ := net.SplitHostPort(listener.Addr().String())

  portInt, err := strconv.Atoi(portStr)

	if err != nil {
		log.Fatalf("Failed to parse port number: %v", err)
	}

	// Convert to int32
	port := int32(portInt)
  makePortMapRequest( subdomain, port )

	log.Printf("Reverse tunnel established on remote port: %v", port)


	for {
		remoteConn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection", err)
			if err == io.EOF {
				return portInt, fmt.Errorf("SSH connection closed")
			}
			continue
		}

		go tun.handleConnection(remoteConn)
	}
}

func makePortMapRequest( subdomain string, port int32 ) {
  conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

  if err != nil {
    fmt.Println( "Error connecting to grpc server ", err.Error() )
  }

  defer conn.Close()

  c := pb.NewTunnelManagerClient(conn)

  ctx, cancel := context.WithTimeout( context.Background(), time.Second )

  defer cancel()

  r, err := c.RequestTunnel( ctx, &pb.RequestTunnelRequest{
    Subdomain: subdomain,
    Port: port,
  })

  if err != nil {
    log.Fatalf("could not create tunnel: %v", err)
  }

  log.Printf("Tunnel created: %v", r.GetPort() )
}

func (tun *ReverseTunnel) handleConnection(remoteConn net.Conn) {
	defer remoteConn.Close()

	localConn, err := net.Dial("tcp", tun.LocalHost)

	if err != nil {
		log.Printf("Failed to connect to local service: %v", err)
		return
	}

	defer localConn.Close()

	go io.Copy(remoteConn, localConn)
	io.Copy(localConn, remoteConn)
}

func getSSHKey(keyPath string) ssh.AuthMethod {
	key, err := os.ReadFile(keyPath)

	if err != nil {
		log.Fatal("Unable to read private key", err.Error())
	}

	signer, err := ssh.ParsePrivateKey(key)

	if err != nil {
		log.Fatal("Unable to parse private key", err.Error())
	}

	return ssh.PublicKeys(signer)
}

func main() {
  subdomain := os.Args[1]

	keyPath := "/Users/kfellows/.ssh/id_rsa"
	authMethod := getSSHKey(keyPath)

	sshConfig := &ssh.ClientConfig{
		User: "kennyfellows",
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	tunnel := &ReverseTunnel{
		RemoteHost: "10.0.0.187",
		RemotePort: 22,
		LocalHost:  "localhost:3333",
		SSHConfig:  sshConfig,
	}

	// Use a channel to signal when the tunnel is ready
	ready := make(chan int)

	// Start the tunnel in a goroutine
	go func() {
		port, err := tunnel.Start( subdomain )
		if err != nil {
			log.Fatalf("Failed to start reverse tunnel: %v", err)
		}
		ready <- port
	}()

	// Wait for the tunnel to be ready and print the port
	port := <-ready
	fmt.Printf("Tunnel established on remote port: %d\n", port)

	// Keep the main goroutine running and handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down...")
}


