package main

import (
  "os"
  "log"
  "fmt"
  "net"
  "bufio"
  "strings"
)

func main() {
  arguments := os.Args
  port := validateArguments( arguments )

  listener, err := net.Listen( "tcp", "localhost:"+port )

  if err != nil {
    log.Fatal( "Error starting server: ", err.Error() )
  }

  defer listener.Close()

  fmt.Printf( "Server lisening on %s", port )

  for {
    conn, err := listener.Accept()

    if err != nil {
      fmt.Println( "Error accepting request: ", err.Error() )
    }

    go handleRequest( conn )
  }
}

func validateArguments( arguments []string ) string {

  if len( arguments ) < 2 {
    log.Fatal("Must provide a port number")
  }

  return arguments[ 1 ]
}

func handleRequest( conn net.Conn ) {
  defer conn.Close()

  reader := bufio.NewReader( conn )

  requestLine, err := reader.ReadString('\n')

  if err != nil {
    fmt.Println( "Error reading request", err )
    return
  }

  parts := strings.Fields( requestLine )

  if len( parts ) < 3 {
    fmt.Println( "Invalid HTTP request" )
    return
  }

  method, path, version := parts[ 0 ], parts[ 1 ], parts[ 2 ]

  fmt.Printf( "Received %s request for %s using %s\n", method, path, version )

  headers, err := parseHeaders( reader )

  if err != nil {
    fmt.Println( "Error parsing headers: ", err )
  }

  fmt.Println( "Parsed headers:" )

  for key, value := range headers {
    fmt.Printf( "%s: %s\n", key, value )
  }

  response := "HTTP/1.1 200 OK\r\n" +
  "Content-Type: text/plain\r\n" +
  "Connection: close\r\n" +
  "\r\n" +
  "Received new request\r\n"

  conn.Write( []byte( response ) )
}

func parseHeaders( reader *bufio.Reader ) ( map[string]string, error ) {
  headers := make( map[string]string )

  for {
    line, err := reader.ReadString('\n')
    if err != nil {
      fmt.Println( "Error reading headers: ", err )
      return nil, err
    }

    line = strings.TrimRight( line, "\r\n" )

    // empty line means we are done with the headers
    if line == "" {
      break
    }

    parts := strings.SplitN( line, ":", 2 )

    if len( parts ) == 2 {
      key := strings.TrimSpace( parts[ 0 ] )
      value := strings.TrimSpace( parts[ 1 ] )
      headers[ key ] = value
    }
  }

  return headers, nil
}
