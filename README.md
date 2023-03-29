English | [中文](./README-CN.md)
<div align="center">

# Shermie-Proxy

</div>

<div align="center">

![GitHub User's stars](https://img.shields.io/github/stars/kxg3030?style=social)
![GitHub go.mod Go version (branch & subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/kxg3030/shermie-proxy/master)
![GitHub](https://img.shields.io/github/license/kxg3030/shermie-proxy)
![GitHub search hit counter](https://img.shields.io/github/search/kxg3030/shermie-proxy/start)
![GitHub release (by tag)](https://img.shields.io/github/downloads/kxg3030/shermie-proxy/v1.1/total)
![GitHub commit activity](https://img.shields.io/github/commit-activity/w/kxg3030/shermie-proxy)
</div>


# Function

- Support http, https, ws, wss, tcp, socket5 protocol data receiving and sending, only need to monitor one port
- Support data interception and custom modification
- Automatically identify inbound protocols, and process them according to different protocols identified by message headers
- Support adding upper-level tcp proxy, only first-level

TODO：

- tcp connection multiplexing

# How to use

- Install
```go
go get github.com/kxg3030/shermie-proxy
```

- Listen
```go
package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"github.com/kxg3030/shermie-proxy/Core"
	"github.com/kxg3030/shermie-proxy/Core/Websocket"
	"github.com/kxg3030/shermie-proxy/Log"
	"io"
	"net/http"
	"strings"
)

func init() {
	Log.NewLogger().Init()
	// Initialize the root certificate
	err := Core.NewCertificate().Init()
	if err != nil {
		return
	}
}

func main() {
	port  := flag.String("port", "9090", "listen port")
	nagle := flag.Bool("nagle", true, "connect remote use nagle algorithm")
	proxy := flag.String("proxy", "", "prxoy remote host")
	to    := flag.String("to", "", "tcp remote host")
	flag.Parse()
	if *port == "0" {
		Log.Log.Fatal("port required")
		return
	}
	// start service
	s := Core.NewProxyServer(*port, *nagle, *proxy, *to)

	// Register http client request event function
	s.OnHttpRequestEvent = func(body []byte, request *http.Request, resolve Core.ResolveHttpRequest, conn net.Conn) {
		mimeType := request.Header.Get("Content-Type")
		if strings.Contains(mimeType, "json") {
			Log.Log.Println("HttpRequestEvent：" + string(body))
		}
		// Data modification can be done here
		resolve(body, request)
	}
	// Register tcp connection event
	s.OnTcpConnectEvent = func(conn net.Conn) {

	}
	// Register tcp close event
	s.OnTcpCloseEvent = func(conn net.Conn) {

	}
	
	s.OnHttpRequestEvent = func(message []byte, request *http.Request, resolve Core.ResolveHttpRequest, conn net.Conn) {
		Log.Log.Println("HttpRequestEvent：" + conn.RemoteAddr().String())
		resolve(message, request)
	}

	s.OnHttpResponseEvent = func(body []byte, response *http.Response, resolve Core.ResolveHttpResponse, conn net.Conn) {
		mimeType := response.Header.Get("Content-Type")
		if strings.Contains(mimeType, "json") {
			Log.Log.Println("HttpResponseEvent：" + string(body))
		}
		// Data modification can be done here
		resolve(body, response)
	}


	s.OnSocks5ResponseEvent = func(message []byte, resolve Core.ResolveSocks5, conn net.Conn) (int, error) {
		Log.Log.Println("Socks5ResponseEvent：" + string(message))
		// Data modification can be done here
		return resolve(message)
	}


	s.OnSocks5RequestEvent = func(message []byte, resolve Core.ResolveSocks5, conn net.Conn) (int, error) {
		Log.Log.Println("Socks5RequestEvent：" + string(message))
		// Data modification can be done here
		return resolve(message)
	}


	s.OnWsRequestEvent = func(msgType int, message []byte, resolve Core.ResolveWs, conn net.Conn) error {
		Log.Log.Println("WsRequestEvent：" + string(message))
		// Data modification can be done here
		return resolve(msgType, message)
	}


	s.OnWsResponseEvent = func(msgType int, message []byte, resolve Core.ResolveWs, conn net.Conn) error {
		Log.Log.Println("WsResponseEvent：" + string(message))
		// Data modification can be done here
		return resolve(msgType, message)
	}


	s.OnTcpClientStreamEvent = func(message []byte, resolve Core.ResolveTcp, conn net.Conn) (int, error) {
		Log.Log.Println("TcpClientStreamEvent：" + string(message))
		// Data modification can be done here
		return resolve(message)
	}


	s.OnTcpServerStreamEvent = func(message []byte, resolve Core.ResolveTcp, conn net.Conn) (int, error) {
		Log.Log.Println("TcpServerStreamEvent：" + string(message))
		// Data modification can be done here
		return resolve(message)
	}

	_ = s.Start()
}
```
- parameter


    --port:The port that the proxy service listens to, the default is 9090


    --to:When proxying tcp service, the ip and port of the destination server, the default is 0, only used by tcp proxy


    --proxy:Superior agent address


    --nagle:Whether to enable the nagle data merging algorithm, the default is true


<div align="center">

<a href="https://t.zsxq.com/0allV9fqi" style="font-size:16px;font-weight:bold">点击加入我的星球</a>
<a href="https://t.zsxq.com/0allV9fqi" style="font-size:16px;font-weight:bold">QQ群：931649621</a>

</div>
<br/>

