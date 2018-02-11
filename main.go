package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

var (
	host        = flag.String("host", "", "IP to listen on")
	port        = flag.Int("port", 514, "port to listen on")
	proto       = flag.String("proto", "tcp", "protocol to use")
	redisServer = flag.String("redisServer", "localhost:6379", "address and port for redis server")
	pool        *redis.Pool
)

type logResponse struct {
	Timestamp string `json:"@timestamp"`
	Message   string `json:"message"`
	Host      string `json:"sysloghost"`
	Severity  string `json:"severity"`
}

var sevs string

func main() {
	flag.Parse()
	sevs = strings.Join(flag.Args(), "")
	if sevs == "" {
		sevs = "emergalertcriterr"
	}

	listener, err := net.Listen(*proto, *host+":"+strconv.Itoa(*port))
	handleErr(err)

	pool = newPool(*redisServer)

	defer listener.Close()
	fmt.Println("Listening on " + *host + ":" + strconv.Itoa(*port))
	for {
		conn, err := listener.Accept()
		handleErr(err)

		go handleRequest(conn, pool.Get())
	}
}

func handleErr(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}

func handleRequest(conn net.Conn, redisConn redis.Conn) {
	defer conn.Close()
	defer redisConn.Close()

	buf := make([]byte, 1024)
	reqLen, err := conn.Read(buf)
	handleErr(err)

	response := buf[0:reqLen]
	data := logResponse{}
	json.Unmarshal(response, &data)

	if strings.Contains(sevs, data.Severity) && data.Severity != "" {
		redisConn.Do("XADD", "logs", "*", "host", data.Host, "timestamp", data.Timestamp, "severity", data.Severity, "message", data.Message)
	}
}