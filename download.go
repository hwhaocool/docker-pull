package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"context"

	"github.com/imroc/req/v3"

	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/pkg/blobinfocache/none"
	"github.com/containers/image/v5/types"

	"github.com/sirupsen/logrus"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func init() {
	// 设置日志级别
	logrus.SetLevel(logrus.DebugLevel)
}

func DownloadImage(cmd Cmd) {

	ParseImageInfo(cmd.image)

	client := req.C()        // Use C() to create a client.
	resp, err := client.R(). // Use R() to create a request.
					Get("https://httpbin.org/uuid")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp)

	main222(cmd)

}

func main222(cmd Cmd) {
	ctx := context.Background()
	sysCtx := &types.SystemContext{}

	// 创建 Docker 引用
	if !strings.HasPrefix(cmd.image, "//") {
		cmd.image = "//" + cmd.image
	}
	ref, err := docker.ParseReference(cmd.image)
	if err != nil {
		log.Fatal(err)
	}

	// 2. 创建镜像源
	src, err := ref.NewImageSource(ctx, sysCtx)
	if err != nil {
		log.Fatalf("Failed to create image source: %v", err)
	}
	defer src.Close()

	// 3. 获取原始 manifest 字节
	rawManifest, _, err := src.GetManifest(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to get manifest: %v", err)
	}

	// 4. 解析 manifest
	digest, err := manifest.Digest(rawManifest)
	if err != nil {
		log.Fatalf("Failed to compute digest: %v", err)
	}

	mediaType := manifest.GuessMIMEType(rawManifest)
	fmt.Printf("Manifest Digest: %s\n", digest)
	fmt.Printf("Media Type: %s\n", mediaType)

	// 5. 根据媒体类型解析具体 manifest
	switch mediaType {
	case manifest.DockerV2Schema2MediaType, ocispec.MediaTypeImageManifest:
		// 解析为 Docker Schema 2 或 OCI Manifest
		var man manifest.Schema2
		if err := json.Unmarshal(rawManifest, &man); err != nil {
			log.Fatalf("Failed to unmarshal manifest: %v", err)
		}
		printSchema2Manifest(man)

	case manifest.DockerV2ListMediaType, ocispec.MediaTypeImageIndex:
		// 解析为 Manifest List
		var list manifest.Schema2List
		if err := json.Unmarshal(rawManifest, &list); err != nil {
			log.Fatalf("Failed to unmarshal manifest list: %v", err)
		}
		printManifestList(list)

		// 转换为 OCI Index
		// ociIndex, err := manifest.OCI1IndexFromManifest(rawManifest)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// convertToOCIIndex(rawManifest)

		downloadWithList(ref, list, src, ctx)

	default:
		log.Fatalf("Unsupported manifest type: %s", mediaType)
	}
}

func printSchema2Manifest(man manifest.Schema2) {
	fmt.Println("\n=== Schema 2 Manifest ===")
	fmt.Printf("Schema Version: %d\n", man.SchemaVersion)
	fmt.Printf("MediaType: %s\n", man.MediaType)
	fmt.Printf("Config:\n")
	fmt.Printf("  MediaType: %s\n", man.ConfigDescriptor.MediaType)
	fmt.Printf("  Size: %d bytes\n", man.ConfigDescriptor.Size)
	fmt.Printf("  Digest: %s\n", man.ConfigDescriptor.Digest)

	fmt.Println("\nLayers:")
	for i, layer := range man.LayerInfos() {
		fmt.Printf("  Layer %d:\n", i+1)
		fmt.Printf("    MediaType: %s\n", layer.MediaType)
		fmt.Printf("    Size: %d bytes\n", layer.Size)
		fmt.Printf("    Digest: %s\n", layer.Digest)
	}
}

func printManifestList(list manifest.Schema2List) {
	fmt.Println("\n=== Manifest List ===")
	fmt.Printf("Schema Version: %d\n", list.SchemaVersion)
	fmt.Printf("MediaType: %s\n", list.MediaType)

	fmt.Println("\nManifests:")
	for i, m := range list.Manifests {
		fmt.Printf("  Manifest %d:\n", i+1)
		fmt.Printf("    MediaType: %s\n", m.MediaType)
		fmt.Printf("    Size: %d bytes\n", m.Size)
		fmt.Printf("    Digest: %s\n", m.Digest)
		// if m.Platform != nil {
		fmt.Printf("    Platform:\n")
		fmt.Printf("      Architecture: %s\n", m.Platform.Architecture)
		fmt.Printf("      OS: %s\n", m.Platform.OS)
		if m.Platform.Variant != "" {
			fmt.Printf("      Variant: %s\n", m.Platform.Variant)
		}
		// }
	}
}

func downloadWithList(ref types.ImageReference, list manifest.Schema2List, src types.ImageSource, ctx context.Context) {

	for _, m := range list.Manifests {

		switch m.MediaType {
		case ocispec.MediaTypeImageManifest:

			if m.Platform.Architecture == "amd64" && m.Platform.OS == "linux" {
				fmt.Printf("  Architecture: %s\n", m.Platform.Architecture)

				log.Println("Downloading manifest for amd64/linux:", m.Digest.String())

				raw, x, err := src.GetManifest(ctx, &m.Digest)
				if err != nil {
					log.Fatal(err)
				}
				log.Println("x:", x)

				// 解析为 Docker Schema 2
				var man manifest.Schema2
				if err := json.Unmarshal(raw, &man); err != nil {
					log.Fatalf("Failed to unmarshal manifest: %v", err)
				}

				downloadConfigBlob(src, man.ConfigDescriptor, ctx)
				downloadLayersBlob(src, man.LayersDescriptors, ctx)

				(&TarInfo{
					Ref:          ref,
					ConfigDigest: strings.TrimPrefix(man.ConfigDescriptor.Digest.String(), "sha256:"),
					LayersDigest: func() []string {
						var layers []string
						for _, layer := range man.LayersDescriptors {
							layers = append(layers, strings.TrimPrefix(layer.Digest.String(), "sha256:"))
						}
						return layers
					}(),
				}).BuildTar()

			}
		}

	}
}

func downloadLayersBlob(src types.ImageSource, schema2Descriptor []manifest.Schema2Descriptor, ctx context.Context) {
	log.Println("Downloading layers blob")
	for _, desc := range schema2Descriptor {
		downloadBlob(ctx, src, desc, SaveProps{
			path: "layers",
			name: "layer.tar",
		})
	}
}

func downloadConfigBlob(src types.ImageSource, configDescriptor manifest.Schema2Descriptor, ctx context.Context) {

	log.Println("Downloading config blob", configDescriptor.Digest.String())
	downloadBlob(ctx, src, configDescriptor, SaveProps{
		path: "config",
		name: "config.json",
	})
}

type SaveProps struct {
	path string
	name string
}

func downloadBlob(ctx context.Context, src types.ImageSource, desc manifest.Schema2Descriptor, saveProps SaveProps) error {
	// 创建 blob 文件夹
	blobPath := filepath.Join("cache", saveProps.path, strings.TrimPrefix(desc.Digest.String(), "sha256:"))

	err := os.MkdirAll(blobPath, 0755)
	if err != nil {
		log.Fatalf("创建目录失败: %v", err)
	}

	// blob 文件
	tarFilePath := filepath.Join(blobPath, saveProps.name)

	// 检查文件是否已存在
	if _, err := os.Stat(tarFilePath); err == nil {
		fmt.Printf("Blob already exists, skipping: %s\n", desc.Digest)
		return nil
	}

	// 创建目标文件
	tarFile, err := os.Create(tarFilePath)
	if err != nil {
		return fmt.Errorf("failed to create blob file: %v", err)
	}
	defer tarFile.Close()

	// 获取 blob 读取器
	blobReader, size, err := src.GetBlob(ctx, types.BlobInfo{
		Digest: desc.Digest,
		Size:   desc.Size,
	}, none.NoCache)

	if err != nil {
		os.Remove(tarFilePath)
		log.Fatal("failed to get blob reader:", err)
		return fmt.Errorf("failed to get blob reader: %v", err)
	}
	defer blobReader.Close()

	// 复制 blob 内容
	copied, err := io.Copy(tarFile, blobReader)
	if err != nil {
		return fmt.Errorf("failed to copy blob content: %v", err)
	}

	// 验证大小
	if copied != size {
		return fmt.Errorf("blob size mismatch: expected %d, got %d", size, copied)
	}

	fmt.Printf("  Successfully downloaded blob: %s (%d bytes)\n", desc.Digest, copied)
	return nil
}
func downloadWithList2(list *manifest.OCI1Index) {

	for _, m := range list.Manifests {

		switch m.MediaType {
		case ocispec.MediaTypeImageManifest:

			if m.Annotations != nil {
				for key, value := range m.Annotations {
					fmt.Printf("  %s: %s\n", key, value)

					// ocispec.Arch
				}
			}
		}

	}
}

func convertToOCIIndex(rawManifest []byte) {
	// 转换为 OCI Index
	ociIndex, err := manifest.OCI1IndexFromManifest(rawManifest)
	if err != nil {
		log.Fatal(err)
	}

	// 现在 ociIndex 是 *ocispec.Index 类型
	fmt.Printf("OCI Index: %+v\n", ociIndex)

	// 访问 annotations
	for i, desc := range ociIndex.Manifests {
		fmt.Printf("\nManifest %d Annotations:\n", i+1)
		if desc.Annotations != nil {
			for key, value := range desc.Annotations {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}
}

type Schema2ListPublic struct {
	SchemaVersion int                         `json:"schemaVersion"`
	MediaType     string                      `json:"mediaType"`
	Manifests     []Schema2ManifestDescriptor `json:"manifests"`
}

type Schema2ManifestDescriptor struct {
	MediaType string              `json:"mediaType"`
	Size      int64               `json:"size"`
	Digest    string              `json:"digest"`
	URLs      []string            `json:"urls,omitempty"`
	Platform  Schema2PlatformSpec `json:"platform"`
}

type Schema2PlatformSpec struct {
	Architecture string   `json:"architecture"`
	OS           string   `json:"os"`
	OSVersion    string   `json:"os.version,omitempty"`
	OSFeatures   []string `json:"os.features,omitempty"`
	Variant      string   `json:"variant,omitempty"`
	Features     []string `json:"features,omitempty"` // removed in OCI
}
