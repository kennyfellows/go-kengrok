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

	"golang.org/x/crypto/ssh"
  "go-kengrok/utils"
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

	log.Printf("Reverse tunnel established on remote port: %v", portInt )

  err = makePortMapRequest( subdomain, portInt )

  if err != nil {
    log.Fatalf( "Failed to start proxy %v", err )
  }

  bindCleanupEvent( subdomain )

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

func makePortMapRequest( subdomain string, port int ) error  {

  ctx, cancel := context.WithTimeout( context.Background(), 3 * time.Second )

  defer cancel()

  pClient, err := utils.GetProxyManagerClient()

  if err != nil {
    return err
  }

  return pClient.StartProxyMapping( ctx, subdomain, port )
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

func cleanupPortMapping( subdomain string ) {
  ctx, cancel := context.WithTimeout( context.Background(), 3 * time.Second )

  defer cancel()

  pClient, err := utils.GetProxyManagerClient()

  if err != nil {
    log.Fatalf( "Error getting proxy manager client: %v", err )
  }

  err = pClient.EndProxyMapping( ctx, subdomain )

  if err != nil {
    log.Fatalf( "Error ending proxy mapping on server: %v",  err )
  }
}

func bindCleanupEvent( subdomain string ) {
  c := make(chan os.Signal, 1)
  signal.Notify(c, os.Interrupt, syscall.SIGTERM)
  go func() {
    <-c
    fmt.Println("\nReceived termination signal")
    cleanupPortMapping( subdomain )
    os.Exit(0)
  }()
}

func main() {

  if len( os.Args ) < 2 {
    log.Fatal("Must provide a local port as first argument")
  }

  if len( os.Args ) < 3 {
    log.Fatal("Must provide a subdomain")
  }

  localPort := os.Args[1]
  subdomain := os.Args[2]

	keyPath := "/Users/kfellows/.ssh/id_rsa"
	authMethod := getSSHKey(keyPath)

	sshConfig := &ssh.ClientConfig{
		User: "kennyfellows",
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

  localDest := fmt.Sprintf( "localhost:%v", localPort )
	tunnel := &ReverseTunnel{
		RemoteHost: "10.0.0.187",
		RemotePort: 22,
		LocalHost:  localDest,
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


