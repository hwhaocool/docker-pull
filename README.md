
## 功能描述
1. 通过 docker regitry v2接口，下载相关的 manifest和layer文件，组装为tar包，可供 docker load使用
2. 无需提前安装`docker`和`podman`等
3. 支持`windows`
4. 程序是二进制包，下载即用，无任何依赖
5. 功能类似 [docker-drag](https://github.com/NotGlop/docker-drag) ，但是无需安装`python`

## 使用说明
```
docker-pull -image 镜像 
```

参数：
-image 镜像名称，支持如 `alpine:3.22.1`; `nginx`; `library/nginx:1.20`; `docker.io/library/nginx:latest`; `myregistry.com/myproject/myapp:v1.0`; `myregistry.com:5000/myproject/myapp:v1.0` 等格式


1. 镜像默认保存到当前目录下的 `output/{namespace}/{repository}`里面
2. layers的缓存目录是当前文件夹下的 `cache`; 其中 `layers`是`layer`， `config`是`config`
3. 组装tar包的时候，会把相关的文件复制到`tmp`目录下
2. 如果 registry 需要鉴权，会自动鉴权

>> 目前仅支持下载 amd64+linux 的镜像，其他架构和系统待支持中

