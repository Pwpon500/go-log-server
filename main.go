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
	maxLen      = flag.Int("maxLen", 1000000, "maximum stream size")
	exactMax    = flag.Bool("exactMax", false, "require exactly maxLen entries")
	isClient    = flag.Bool("c", false, "run as client")
	pool        *redis.Pool
	baseArgs    = redis.Args{}
	sevs        string
)

type logResponse struct {
	Host      string `json:"sysloghost"`
	Timestamp string `json:"@timestamp"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
}

func main() {
	flag.Parse()

	if *isClient {
		client()
	} else {
		server()
	}
}

func client() {

}

func server() {
	sevs = strings.Join(flag.Args(), "")
	if sevs == "" {
		sevs = "emergalertcriterr"
	}

	listener, err := net.Listen(*proto, *host+":"+strconv.Itoa(*port))
	handleErr(err)
	baseArgs = baseArgs.Add("logs")
	if *maxLen != 0 {
		baseArgs = baseArgs.Add("MAXLEN")
		if !*exactMax {
			baseArgs = baseArgs.Add("~")
		}
		baseArgs = baseArgs.Add(strconv.Itoa(*maxLen))
	}
	baseArgs = baseArgs.Add("*")

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
		redisConn.Do("XADD", baseArgs.AddFlat(data)...)
	}
}
