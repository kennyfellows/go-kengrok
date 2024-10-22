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

	"go-kengrok/utils"

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
	port := 0
	fmt.Sscanf(portStr, "%d", &port)

  savePortMapping( subdomain, port )

	log.Printf("Reverse tunnel established on remote port: %v", port)


	for {
		remoteConn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection", err)
			if err == io.EOF {
				return port, fmt.Errorf("SSH connection closed")
			}
			continue
		}

		go tun.handleConnection(remoteConn)
	}
}

func savePortMapping( subdomain string, port int ) ( ok bool, err error ) {
  redisClient := utils.GetRedisClient()

  ctx := context.Background()
  key := fmt.Sprintf( "kengrok-map:%s", subdomain )
  redisClient.Set( ctx, key, port, 0 ).Err()

  return true, nil
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

  isInUse := subdomainInUse( subdomain )

  if isInUse {
    log.Fatalf( "Subdomain %s is already in use", subdomain )
  }

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
