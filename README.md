
## 功能描述
1. 通过 docker regitry v2接口，下载相关的 manifest和layer文件，组装为tar包，可供 docker load使用
2. 无需提前安装`docker`和`podman`等
3. 支持`windows`
4. 程序是二进制包，下载即用，无任何依赖
5. 功能类似 [docker-drag](https://github.com/NotGlop/docker-drag) ，但是无需安装`python`

## 使用说明
```
docker-pull -arch amd64/arm64 -image 镜像 
```

参数：

| 参数 | 说明 | 默认值 | 可选值/格式 |
|-----|-----|------|-------------|
| `-image` | 镜像名称 | 无默认值<br>必填 | `alpine:3.22.1`<br>`nginx`<br>`library/nginx:1.20`<br>`docker.io/library/nginx:latest`<br>`myregistry.com/myproject/myapp:v1.0`<br>`myregistry.com:5000/myproject/myapp:v1.0` |
| `-arch` | 架构 | `amd64` | `amd64` / `arm64` |
| `-proxy` | 代理设置 | 无代理 | `socks5://ip:port`<br>`http://ip:port`<br>`https://ip:port`<br>`http://username:password@ip:port (鉴权格式)` |



1. 镜像默认保存到当前目录下的 `output/{namespace}/{repository}`里面
2. 有一个缓存目录是 `cache`; 其中 由`layer`，和 `config`，在多次下载的时候可以加速；如果觉得占用磁盘可以手动删除，不影响功能
3. 组装tar包的时候，会把相关的文件复制到`tmp`目录下
4. 如果 registry 需要鉴权，会自动鉴权
5. 如果失败，可以反复尝试（下载过程中，如果成功，文件会保留，下次跳过；如果失败，cache会删除）


## 目录说明
1. cache 缓存，包括confi和layer
2. output 输出
3. tmp 临时目录
