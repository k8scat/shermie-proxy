package Core

import (
	"fmt"
	"runtime"

	"github.com/k8scat/shermie-proxy/Log"
)

func (i *ProxyServer) Install() {
	if runtime.GOOS == "windows" {
		err := Utils.InstallCert("cert.crt")
		if err != nil {
			Log.Log.Println(err.Error())
			return
		}
		Log.Log.Println("已安装系统证书")
		err = Utils.SetSystemProxy(fmt.Sprintf("127.0.0.1:%s", i.port))
		if err != nil {
			Log.Log.Println(err.Error())
			return
		}
		Log.Log.Println("已设置系统代理")
		return
	}
	Log.Log.Println("非windows系统请手动安装证书并设置代理,可以在根目录或访问http://127.0.0.1/tls获取证书文件")
}

func (i *ProxyServer) UnInstall() {
	if runtime.GOOS == "windows" {
		err := Utils.SetSystemProxy("")
		if err != nil {
			Log.Log.Println(err.Error())
			return
		}
		Log.Log.Println("已关闭系统代理")
		return
	}
}
