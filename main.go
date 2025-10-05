package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/fatih/color"
)

var Version = "dev"

func main() {
	color.HiMagenta("docker-pull version: %s", Version)
	var image, proxyAddr, destination, arch string

	flag.StringVar(&image, "image", "", "镜像名称, 支持如 alpine:3.22.1; nginx; library/nginx:1.20; docker.io/library/nginx:latest; myregistry.com/myproject/myapp:v1.0; myregistry.com:5000/myproject/myapp:v1.0 等格式")

	flag.StringVar(&arch, "arch", "amd64", "cpu架构, 可选 amd64, arm64, 默认 amd64")

	flag.StringVar(&proxyAddr, "proxy", "", "代理地址, 支持 socks5://username:password@ip:port 或者 http://username:password@ip:port; 也支持不带认证的代理，如 socks5://ip:port 或者 http://ip:port")

	// TODO: 待支持
	// flag.StringVar(&destination, "dst", "output", "镜像保存路径")

	flag.Parse()

	if image == "" {
		log.Fatal("必须提供 -image 参数")
	}

	var proxyURL *url.URL
	var err error
	if proxyAddr != "" {
		// 解析 proxy URL
		proxyURL, err = url.Parse(proxyAddr)
		if err != nil {
			log.Fatal("proxy参数格式错误", err)
		}
	}

	cmd := Cmd{
		image:       image,
		proxy:       proxyURL,
		destination: destination,
		arch:        arch,
	}

	// 确保目标目录存在
	// err := os.MkdirAll(destination, 0755)
	// if err != nil {
	// 	log.Fatalf("创建目录失败: %v", err)
	// }

	DownloadImage(cmd)

	fmt.Println("ok")
}

type Cmd struct {
	image       string
	proxy       *url.URL
	destination string
	arch        string // cpu架构, 可选 amd64, arm64
}
