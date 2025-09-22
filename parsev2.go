package main

import (
	"strings"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/docker/reference"
	"github.com/containers/image/v5/types"
)

type DockerImageV2 struct {
	Domain string
	Path   string
	Tag    string

	Namespace  string
	Repository string
}

func ParseImageInfoV2(image string) (types.ImageReference, DockerImageV2, error) {

	tag := "latest"
	ns := "library"
	repo := ""

	ref, err := docker.ParseReference(image)
	if err != nil {
		return nil, DockerImageV2{}, err
	}

	domain := reference.Domain(ref.DockerReference())
	path := reference.Path(ref.DockerReference())

	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		repo = parts[0]
	} else {
		ns = parts[0]
		repo = strings.Join(parts[1:], "/")
	}

	if tagged, ok := ref.DockerReference().(reference.Tagged); ok {
		tag = tagged.Tag()
	}

	return ref, DockerImageV2{
		Domain:     domain,
		Path:       path,
		Tag:        tag,
		Namespace:  ns,
		Repository: repo,
	}, nil
}
