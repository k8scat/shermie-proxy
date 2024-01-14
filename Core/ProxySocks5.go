package Core

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/k8scat/shermie-proxy/Log"
)

type ProxySocks5 struct {
	ConnPeer
	target net.Conn
	port   string
}

type ResolveSocks5 func(buff []byte) (int, error)

const (
	// 预留位
	Rsv = 0x00
	// 命令
	CommandConn = 0x01
	CommandBind = 0x02
	CommandUdp  = 0x03
	// 目标类型
	TargetIpv4   = 0x01
	TargetIpv6   = 0x04
	TargetDomain = 0x03
	Version      = 0x5
)

const (
	GssApi                = 0x01
	UsernamePassword      = 0x02
	IanaAssignedMin       = 0x03
	IanaAssignedMax       = 0x7F
	ReservedForPrivateMin = 0x80
	ReservedForPrivateMax = 0xFE
	NoAcceptMethod        = 0xFF
)

const SocketServer = "server"
const SocketClient = "client"

func NewProxySocket() *ProxySocks5 {
	return &ProxySocks5{}
}

func (i *ProxySocks5) Handle() {
	// 读取版本号
	version, err := i.reader.ReadByte()
	if err != nil {
		Log.Log.Println("读取socks5版本号错误：" + err.Error())
		return
	}
	if version != Version {
		Log.Log.Println("socks5版本号不匹配")
		return
	}
	// 读取支持的方法
	methodNum, err := i.reader.ReadByte()
	if err != nil {
		Log.Log.Println("读取socks5支持方法数量错误：" + err.Error())
		return
	}
	if methodNum < 0 || methodNum > 0xFF {
		Log.Log.Println("socks5支持方法参数错误")
		return
	}
	// 代理默认不需要账号密码验证
	var requiredAuth bool = false
	// 读取所有的方法列表
	for n := 0; n < int(methodNum); n++ {
		_, err := i.reader.ReadByte()
		if err != nil {
			Log.Log.Println("读取socks5支持错误：" + err.Error())
			return
		}
	}
	_, err = i.writer.Write([]byte{Version, 0x00})
	if err != nil {
		Log.Log.Println("返回数据错误：" + err.Error())
		return
	}
	_ = i.writer.Flush()
	if requiredAuth {
		// TODO 账号密码验证
		return
	}
	// 读取版本号
	version, err = i.reader.ReadByte()
	if version != Version {
		Log.Log.Println("socks5版本号错误")
		return
	}
	// 读取命令
	command, err := i.reader.ReadByte()
	if err != nil {
		Log.Log.Println("读取socks5命令错误")
		return
	}
	if command != CommandConn && command != CommandBind && command != CommandUdp {
		Log.Log.Println("不支持socks5命令")
		return
	}
	// 读取保留位
	rsv, err := i.reader.ReadByte()
	if err != nil || rsv != Rsv {
		Log.Log.Println("读取socks5保留位错误")
		return
	}
	// 读取目标地址类型
	targetType, err := i.reader.ReadByte()
	if err != nil {
		Log.Log.Println("读取socks5保留位错误")
		return
	}
	if targetType != TargetIpv4 && targetType != TargetIpv6 && targetType != TargetDomain {
		Log.Log.Println("不支持socks5地址")
		return
	}
	var hostname string
	switch targetType {
	case TargetIpv4:
		buffer := make([]byte, 4)
		// 读4字节
		n, err := i.reader.Read(buffer)
		if err != nil || n != len(buffer) {
			Log.Log.Println("读取ipv4地址错误")
			return
		}
		hostname = net.IP(buffer).String()
		break
	case TargetIpv6:
		buffer := make([]byte, 16)
		// 读16字节
		n, err := i.reader.Read(buffer)
		if err != nil || n != len(buffer) {
			Log.Log.Println("读取ipv6地址错误")
			return
		}
		hostname = net.IP(buffer).String()
		break
	case TargetDomain:
		// 读取域名长度
		domainLen, err := i.reader.ReadByte()
		if err != nil || domainLen <= 0 {
			Log.Log.Println("读取域名地址错误")
			return
		}
		buffer := make([]byte, domainLen)
		n, err := i.reader.Read(buffer)
		if err != nil || n != len(buffer) {
			Log.Log.Println("读取域名地址错误")
			return
		}
		addr, err := net.ResolveIPAddr("ip", string(buffer))
		if err != nil {
			Log.Log.Println("读取域名地址错误：" + err.Error())
			hostname = string(buffer)
		} else {
			hostname = addr.String()
		}
		break
	}
	// 读端口号,大端
	buffer := make([]byte, 2)
	_, err = i.reader.Read(buffer)
	if err != nil {
		Log.Log.Println("读取端口号错误：" + err.Error())
		return
	}
	i.port = strconv.Itoa(int(i.ByteToInt(buffer)))
	hostname = fmt.Sprintf("%s:%s", hostname, i.port)
	// 写入版本号
	_ = i.writer.WriteByte(Version)
	if command == CommandUdp {
		i.target, err = net.DialTimeout("udp", hostname, time.Second*30)
	} else {
		if i.port == "443" {
			dialer := &net.Dialer{
				Timeout: time.Second * 30,
			}
			i.target, err = tls.DialWithDialer(dialer, "tcp", hostname, &tls.Config{
				InsecureSkipVerify: true,
			})
		} else {
			i.target, err = net.DialTimeout("tcp", hostname, time.Second*30)
		}
	}
	Log.Log.Println("待连接的目标服务器：" + hostname)
	// 写入Rep
	if err != nil {
		Log.Log.Println("连接目标服务器失败：" + hostname + " " + err.Error())
		_ = i.writer.WriteByte(0x01)
		_ = i.writer.Flush()
		return
	} else {
		_ = i.writer.WriteByte(0x00)
	}
	defer func() {
		i.target.Close()
	}()
	// 写入Rsv
	_ = i.writer.WriteByte(Rsv)
	remoteAddr := i.target.RemoteAddr().String()
	host, _, _ := net.SplitHostPort(remoteAddr)
	if i.IpV4(host) {
		_ = i.writer.WriteByte(TargetIpv4)
		_, _ = i.writer.Write(net.ParseIP(host).To4())
	}
	if i.IpV6(host) {
		_ = i.writer.WriteByte(TargetIpv6)
		_, _ = i.writer.Write(net.ParseIP(host).To16())
	}
	if !i.IpV4(host) && !i.IpV6(host) {
		_ = i.writer.WriteByte(TargetDomain)
		_ = i.writer.WriteByte(byte(len(hostname)))
		_, _ = i.writer.WriteString(hostname)
	}
	// 写入端口
	_, _ = i.writer.Write(buffer)
	err = i.writer.Flush()
	if err != nil {
		Log.Log.Println("写入socks5握手错误：" + err.Error())
		return
	}
	out := make(chan error, 1)
	if command == 0x01 {
		go i.Transport(out, i.conn, i.target, SocketClient)
		go i.Transport(out, i.target, i.conn, SocketServer)
	}
	err = <-out
	Log.Log.Println("代理socks5数据错误：" + err.Error())
}

func (i *ProxySocks5) Transport(out chan<- error, originConn net.Conn, targetConn net.Conn, role string) {
	buff := make([]byte, 10*1024)
	resolve := ResolveSocks5(func(buff []byte) (int, error) {
		return targetConn.Write(buff)
	})

	var writeLen int
	for {
		readLen, err := originConn.Read(buff)
		if readLen > 0 {
			buff = buff[0:readLen]
			if role == SocketServer {
				if i.server.OnSocks5ResponseEvent != nil {
					writeLen, err = i.server.OnSocks5ResponseEvent(buff, resolve, i.conn)
				} else {
					writeLen, err = resolve(buff)
				}
			} else {
				if i.server.OnSocks5RequestEvent != nil {
					writeLen, err = i.server.OnSocks5RequestEvent(buff, resolve, i.conn)
				} else {
					writeLen, err = resolve(buff)
				}
			}
			if writeLen < 0 || readLen < writeLen {
				writeLen = 0
				if err == nil {
					out <- errors.New("写入目标服务器错误-1")
					break
				}
			}
			if readLen != writeLen {
				out <- errors.New("写入目标服务器错误-2")
				break
			}
		}
		if err != nil {
			out <- errors.New("读取客户端数据错误-1")
			break
		}
		buff = buff[:]
	}
}

func (i *ProxySocks5) IpV4(ipAddr string) bool {
	ip := net.ParseIP(ipAddr)
	return ip != nil && strings.Contains(ipAddr, ".")
}

func (i *ProxySocks5) IpV6(ipAddr string) bool {
	ip := net.ParseIP(ipAddr)
	return ip != nil && strings.Contains(ipAddr, ":")
}

// 字节转整型
func (i *ProxySocks5) ByteToInt(input []byte) int32 {
	return int32(input[0]&0xFF)<<8 | int32(input[1]&0xFF)
}
