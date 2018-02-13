# go-log-server

go-log-server is a Go-based log collector. It writes the logs it collects to a Redis stream.

Logs must be formatted in JSON in this rsyslog format:
```
template(name="json-template"
  type="list") {
    constant(value="{")
      constant(value="\"@timestamp\":\"")     property(name="timereported" dateFormat="rfc3339")
      constant(value="\",\"@version\":\"1")
      constant(value="\",\"message\":\"")     property(name="msg" format="json")
      constant(value="\",\"sysloghost\":\"")  property(name="hostname")
      constant(value="\",\"severity\":\"")    property(name="syslogseverity-text")
      constant(value="\",\"facility\":\"")    property(name="syslogfacility-text")
      constant(value="\",\"programname\":\"") property(name="programname")
      constant(value="\",\"procid\":\"")      property(name="procid")
    constant(value="\"}\n")
}
```

go-log-server uses TCP as the transport protocol for logs. In rsyslog, TCP messages can be sent by following this format:
````
## send to syslog server
*.*                         @@server_ip:514;json-template
```

## Usage

The command-line options can be listed using the `-h` flag:
```
$ go-log-server -h
Usage of ./go-log-server:
  -c	run as client
  -exactMax
        require exactly maxLen entries
  -host string
        IP to listen on
  -maxLen int
        maximum stream size (default 1000000)
  -port int
        port to listen on (default 514)
  -proto string
        protocol to use (default "tcp")
  -redisServer string
        address and port for redis server (default "localhost:6379")
```

