package Core

import (
	"github.com/k8scat/shermie-proxy/Log"
)

func (i *ProxyServer) Install() {
	Log.Log.Println("非windows系统请手动安装证书并设置代理,可以在根目录或访问http://127.0.0.1/tls获取证书文件")
}

func (i *ProxyServer) UnInstall() {
	//
}
