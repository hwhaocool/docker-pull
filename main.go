package main

import (
	"flag"
	"fmt"
	"log"

	"os"
)

func main() {
	var image, proxyAddr, destination string

	flag.StringVar(&image, "image", "demo.52013120.xyz/alpine:3.22.1", "镜像名称")
	flag.StringVar(&proxyAddr, "proxy", "", "socks5代理地址")
	flag.StringVar(&destination, "dst", "output", "镜像保存路径")

	flag.Parse()

	if image == "" {
		log.Fatal("必须提供 -image 参数")
	}

	cmd := Cmd{
		image:       image,
		proxyAddr:   proxyAddr,
		destination: destination,
	}

	// 确保目标目录存在
	err := os.MkdirAll(destination, 0755)
	if err != nil {
		log.Fatalf("创建目录失败: %v", err)
	}

	DownloadImage(cmd)

	// 构建镜像下载URL (简化实现)
	// 在实际的docker镜像下载中，这会更复杂，需要与Docker Registry API交互
	// imageURL := fmt.Sprintf("https://registry-1.docker.io/v2/%s/manifests/latest", image)

	// // 创建HTTP客户端
	// client := &http.Client{}

	// // 如果提供了代理地址，则配置SOCKS5代理
	// if proxyAddr != "" {
	// 	dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
	// 	if err != nil {
	// 		log.Fatalf("无法连接到SOCKS5代理: %v", err)
	// 	}

	// 	httpTransport := &http.Transport{
	// 		Dial: dialer.Dial,
	// 	}
	// 	client.Transport = httpTransport
	// }

	// // 发送请求下载镜像
	// resp, err := client.Get(imageURL)
	// if err != nil {
	// 	log.Fatalf("下载镜像失败: %v", err)
	// }
	// defer resp.Body.Close()

	// if resp.StatusCode != http.StatusOK {
	// 	log.Fatalf("下载镜像失败，状态码: %d", resp.StatusCode)
	// }

	// // 创建目标文件
	// filename := filepath.Join(destination, fmt.Sprintf("%s_manifest.json", image))
	// file, err := os.Create(filename)
	// if err != nil {
	// 	log.Fatalf("创建文件失败: %v", err)
	// }
	// defer file.Close()

	// // 将响应内容写入文件
	// _, err = file.ReadFrom(resp.Body)
	// if err != nil {
	// 	log.Fatalf("写入文件失败: %v", err)
	// }

	// fmt.Printf("镜像 %s 已保存到 %s\n", image, filename)

	fmt.Println("ok")
}

type Cmd struct {
	image       string
	proxyAddr   string
	destination string
}
