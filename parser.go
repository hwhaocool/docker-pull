package main

import (
	"strings"
)

type DockerImage struct {
	ID         int
	Url        string
	Registry   string
	Namespace  string
	Repository string
	Tag        string
	Size       string
	Status     string
	CreateTime string
	UpdateTime string

	official bool
	dockerio bool
}

func ParseImageInfo(image string) (*DockerImage, error) {

	// # 处理 registry/namespace/repository:tag 格式

	// 默认值
	registry := "registry-1.docker.io"
	namespace := "library"
	tag := "latest"

	firstIndex := strings.Index(image, "/")

	if strings.Index(image, ".") < firstIndex {
		// . 在前面，说明是第三方镜像
		registry = image[:firstIndex]

		if registry == "docker.io" {
			// docker.io
			registry = "registry-1.docker.io"
		}
		image = image[firstIndex+1:]
	}

	// 解析 tag
	ss := strings.Split(image, ":")
	if len(ss) > 1 {
		tag = ss[1]
	}

	ss2 := strings.Split(ss[0], "/")

	repository := ss2[len(ss2)-1]

	if len(ss2) == 1 {
		return &DockerImage{
			Registry:   registry,
			Namespace:  namespace,
			Repository: repository,
			Tag:        tag,
			Url:        image,

			official: registry == "docker.io" && namespace == "library",
			dockerio: registry == "docker.io",
		}, nil
	}

	// 切掉最后的一个
	ss2 = ss2[:len(ss2)-1]

	namespace = strings.Join(ss2, "/")

	return &DockerImage{
		Registry:   registry,
		Namespace:  namespace,
		Repository: repository,
		Tag:        tag,
		Url:        image,

		official: registry == "docker.io" && namespace == "library",
		dockerio: registry == "docker.io",
	}, nil

}
