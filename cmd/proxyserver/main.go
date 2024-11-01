package main

import (
  "bufio"
  "bytes"
  "database/sql"
  "fmt"
  "io"
  "log"
  "net"
  "os"
  "strings"
  "time"

  _ "github.com/mattn/go-sqlite3"
)

type Server struct {
  db  *sql.DB
  lis *net.Listener
}

func main() {
  arguments := os.Args
  port := validateArguments(arguments)

  listener, err := net.Listen("tcp", "0.0.0.0:"+port)
  if err != nil {
    log.Fatal("Error starting server: ", err.Error())
  }

  db, err := newSQLiteDB()
  if err != nil {
    log.Fatal("Error connecting to database: ", err.Error())
  }

  server := &Server{
    db:  db,
    lis: &listener,
  }

  defer listener.Close()
  defer db.Close()

  fmt.Printf("Server listening on %s", port)

  for {
    conn, err := listener.Accept()
    if err != nil {
      fmt.Println("Error accepting request: ", err.Error())
    }

    go server.handleRequest(conn)
  }
}

func ParseSubdomain(host string) (string, error) {
  if i := strings.Index(host, ":"); i != -1 {
    host = host[:i]
  }

  parts := strings.Split(host, ".")

  if len(parts) < 2 {
    return "", nil
  }

  subdomain := parts[0]

  return subdomain, nil
}

func newSQLiteDB() (*sql.DB, error) {
  dbPath := os.Getenv("DB_PATH")
  if dbPath == "" {
    dbPath = "/app/data/kengrok.db"
  }

  db, err := sql.Open("sqlite3", dbPath)
  if err != nil {
    return nil, fmt.Errorf("error opening database: %v", err)
  }

  // Test the connection
  err = db.Ping()
  if err != nil {
    return nil, fmt.Errorf("error connecting to database: %v", err)
  }

  return db, nil
}

func (s *Server) getPortMapping(subdomain string) (int, error) {
  var port int
  err := s.db.QueryRow("SELECT proxy_port FROM port_mappings WHERE subdomain = ?", subdomain).Scan(&port)
  if err != nil {
    if err == sql.ErrNoRows {
      return 0, fmt.Errorf("no mapping found for subdomain: %s", subdomain)
    }
    return 0, fmt.Errorf("error querying database: %v", err)
  }

  return port, nil
}

func validateArguments(arguments []string) string {
  if len(arguments) < 2 {
    log.Fatal("Must provide a port number")
  }

  return arguments[1]
}

func (s *Server) proxyRequest(sourceConn net.Conn, dstConn net.Conn) {
  defer dstConn.Close()
  defer sourceConn.Close()

  go io.Copy(dstConn, sourceConn)
  io.Copy(sourceConn, dstConn)
}

func (s *Server) handleRequest(conn net.Conn) {
  defer conn.Close()

  // Set a read deadline to prevent hanging on slow clients
  conn.SetReadDeadline(time.Now().Add(10 * time.Second))

  reader := bufio.NewReader(conn)

  fmt.Println("Parsing the headers")

  headers, headerBytes, err := parseHeaders(reader)

  fmt.Println("Done Parsing the headers")

  if err != nil {
    fmt.Printf("Error parsing headers: %v\n", err)
    return
  }

  subdomain, err := ParseSubdomain(headers["Host"])

  fmt.Printf("\nSubdomain: %s\n", subdomain)

  if err != nil {
    fmt.Printf("Error parsing subdomain: %v\n", err)
  }

  if subdomain == "" {
    fmt.Printf("No subdomain mapping found for '%v'", subdomain)
    return
  }

  // Reset the connection's read deadline for the proxy operation
  conn.SetReadDeadline(time.Time{})

  proxyPort, err := s.getPortMapping(subdomain)

  if err != nil {
    fmt.Printf("Error finding port mapping for subdomain: %v", err)
    return
  }

  fmt.Printf("Subdomain '%v' is mapped to port '%v'", subdomain, proxyPort)
  dstStr := fmt.Sprintf("host.docker.internal:%v", proxyPort)
  dstConn, err := net.Dial("tcp", dstStr)

  if err != nil {
    fmt.Printf("Error establishing proxy connection: %v", err)
    sendBadRequestResponse(conn, "Unable to establish proxy connection")
    return
  }

  _, err = dstConn.Write(headerBytes)

  s.proxyRequest(conn, dstConn)
}

func sendBadRequestResponse(conn net.Conn, msg string) {
  errorMessage := fmt.Sprintf("Bad request: %v", msg)

  response := fmt.Sprintf(
    "HTTP/1.1 400 Bad Request\r\n"+
    "Content-Type: text/plain\r\n"+
    "Content-Length: %d\r\n"+
    "Connection: close\r\n"+
    "\r\n%s",
    len(errorMessage),
    errorMessage,
  )

  conn.Write([]byte(response))
}

func parseHeaders(reader *bufio.Reader) (map[string]string, []byte, error) {
  headers := make(map[string]string)
  var headerBytes []byte

  // Read headers line by line
  for {
    line, err := reader.ReadBytes('\n')
    if err != nil {
      if err == io.EOF {
        return nil, nil, fmt.Errorf("connection closed before headers complete")
      }
      return nil, nil, err
    }

    headerBytes = append(headerBytes, line...)

    trimmed := bytes.TrimRight(line, "\r\n")

    if len(trimmed) == 0 {
      break
    }

    parts := bytes.SplitN(trimmed, []byte(":"), 2)
    if len(parts) == 2 {
      key := strings.TrimSpace(string(parts[0]))
      value := strings.TrimSpace(string(parts[1]))
      headers[key] = value
    }
  }

  return headers, headerBytes, nil
}
