package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/fatih/color"
)

var Version = "dev"

func main() {
	color.HiMagenta("docker-pull version: %s", Version)
	var image, proxyAddr, destination, arch string

	flag.StringVar(&image, "image", "", "镜像名称, 支持如 alpine:3.22.1; nginx; library/nginx:1.20; docker.io/library/nginx:latest; myregistry.com/myproject/myapp:v1.0; myregistry.com:5000/myproject/myapp:v1.0 等格式")

	flag.StringVar(&arch, "arch", "amd64", "cpu架构, 可选 amd64, arm64, 默认 amd64")

	// TODO: 以后支持代理下载
	// flag.StringVar(&proxyAddr, "proxy", "", "socks5代理地址")
	// flag.StringVar(&destination, "dst", "output", "镜像保存路径")

	flag.Parse()

	if image == "" {
		log.Fatal("必须提供 -image 参数")
	}

	cmd := Cmd{
		image:       image,
		proxyAddr:   proxyAddr,
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
	proxyAddr   string
	destination string
	arch        string
}
