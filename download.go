package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"context"

	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/pkg/blobinfocache/none"
	"github.com/containers/image/v5/types"

	"github.com/fatih/color"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func DownloadImage(cmd Cmd) {

	ctx := context.Background()
	sysCtx := &types.SystemContext{
		DockerProxyURL: cmd.proxy, // 设置代理
	}

	// 创建 Docker 引用
	if !strings.HasPrefix(cmd.image, "//") {
		cmd.image = "//" + cmd.image
	}

	ref, imageinfo, err := ParseImageInfoV2(cmd.image)
	if err != nil {
		Logger.Fatal(err)
	}
	printImageInfo(imageinfo)

	// 2. 创建镜像源
	src, err := ref.NewImageSource(ctx, sysCtx)
	if err != nil {
		Logger.Fatalf("Failed to create image source: %v", err)
	}
	defer src.Close()

	// 3. 获取原始 manifest 字节
	rawManifest, _, err := src.GetManifest(ctx, nil)
	if err != nil {
		Logger.Fatalf("Failed to get manifest: %v", err)
	}

	// 4. 解析 manifest
	digest, err := manifest.Digest(rawManifest)
	if err != nil {
		Logger.Fatalf("Failed to compute digest: %v", err)
	}

	mediaType := manifest.GuessMIMEType(rawManifest)
	Logger.Infoln("Manifest Digest:", digest)
	Logger.Infoln("Media Type:", mediaType)

	// 5. 根据媒体类型解析具体 manifest
	switch mediaType {
	case manifest.DockerV2Schema2MediaType, ocispec.MediaTypeImageManifest:
		// 解析为 Docker Schema 2 或 OCI Manifest
		var man manifest.Schema2
		if err := json.Unmarshal(rawManifest, &man); err != nil {
			Logger.Fatalf("Failed to unmarshal manifest: %v", err)
		}
		// printSchema2Manifest(man)

	case manifest.DockerV2ListMediaType, ocispec.MediaTypeImageIndex:
		// 解析为 Manifest List
		var list manifest.Schema2List
		if err := json.Unmarshal(rawManifest, &list); err != nil {
			Logger.Fatalf("Failed to unmarshal manifest list: %v", err)
		}
		// printManifestList(list)

		// 下载
		d := &Downloader{
			ref:         ref,
			src:         src,
			ctx:         ctx,
			schema2List: list,
			imageInfo:   imageinfo,
			cmd:         cmd,
		}
		d.downloadWithList()

	default:
		Logger.Fatalf("Unsupported manifest type: %s", mediaType)
	}
}

func printImageInfo(info DockerImageV2) {
	color.HiCyan("Image Info:")
	color.HiCyan("  Domain: %s", info.Domain)
	color.HiCyan("  Path: %s", info.Path)
	color.HiCyan("  Tag: %s", info.Tag)
	color.HiCyan("  Name: %s", info.Name)
	color.HiCyan("  Namespace: %s", info.Namespace)
	color.HiCyan("  Repository: %s", info.Repository)
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

type Downloader struct {
	ref         types.ImageReference
	src         types.ImageSource
	ctx         context.Context
	schema2List manifest.Schema2List
	imageInfo   DockerImageV2

	cmd Cmd
}

func (d *Downloader) downloadWithList() {

	for _, m := range d.schema2List.Manifests {

		switch m.MediaType {
		case ocispec.MediaTypeImageManifest:

			if m.Platform.Architecture == d.cmd.arch && m.Platform.OS == "linux" {

				Logger.Infof("Downloading manifest for %s/linux: %s\n", d.cmd.arch, m.Digest.String())

				// raw是字节数组， 第二个是 content type
				raw, _, err := d.src.GetManifest(d.ctx, &m.Digest)
				if err != nil {
					Logger.Fatal(err)
				}

				// 解析为 Docker Schema 2
				var man manifest.Schema2
				if err := json.Unmarshal(raw, &man); err != nil {
					Logger.Fatalf("Failed to unmarshal manifest: %v", err)
				}

				var wg sync.WaitGroup
				errChan := make(chan error, len(man.LayersDescriptors)+1)

				// 下载 config
				d.downloadConfigBlob(man.ConfigDescriptor, &wg, errChan)

				// 下载 layers
				d.downloadLayersBlob(man.LayersDescriptors, &wg, errChan)

				wg.Wait()

				// 收集所有错误
				close(errChan)
				hasError := false
				for err := range errChan {
					if err != nil {
						Logger.Errorf("Error occurred: %v", err)
						hasError = true
					}
				}
				// 如果有错误，则退出程序
				if hasError {
					Logger.Fatal("Errors occurred during download, see logs above.")
				}

				// 构造tar包
				(&TarInfo{
					Ref:          d.ref,
					ImageInfo:    d.imageInfo,
					ConfigDigest: strings.TrimPrefix(man.ConfigDescriptor.Digest.String(), "sha256:"),
					Arch:         d.cmd.arch,
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

func (d *Downloader) downloadLayersBlob(schema2Descriptor []manifest.Schema2Descriptor, wg *sync.WaitGroup, errChan chan error) {
	for _, desc := range schema2Descriptor {
		Logger.Info(color.HiCyanString("Downloading layers %s", strings.TrimPrefix(desc.Digest.String(), "sha256:")[:16]))

		wg.Add(1)
		go func(errChan chan error) {
			defer wg.Done()

			blobPath, err := d.downloadBlob(desc, SaveProps{
				path: "layers",
				name: "layer.tar",
			})
			if err != nil {
				errChan <- err

				err2 := os.RemoveAll(blobPath)
				if err2 != nil {
					Logger.Infof("Failed to remove blob file: %v\n", err2)
				}
			}
		}(errChan)
	}
}

func (d *Downloader) downloadConfigBlob(configDescriptor manifest.Schema2Descriptor, wg *sync.WaitGroup, errChan chan error) {

	Logger.Info(color.HiCyanString("Downloading config %s", strings.TrimPrefix(configDescriptor.Digest.String(), "sha256:")[:16]))

	wg.Add(1)
	go func(errChan chan error) {
		defer wg.Done()
		blobPath, err := d.downloadBlob(configDescriptor, SaveProps{
			path: "config",
			name: "config.json",
		})
		if err != nil {

			errChan <- err

			err2 := os.RemoveAll(blobPath)
			if err2 != nil {
				Logger.Infof("Failed to remove blob file: %v\n", err2)
			}
		}
	}(errChan)
}

type SaveProps struct {
	path string
	name string
}

func (d *Downloader) downloadBlob(desc manifest.Schema2Descriptor, saveProps SaveProps) (string, error) {

	// 创建 blob 文件夹
	blobPath := filepath.Join("cache", saveProps.path, strings.TrimPrefix(desc.Digest.String(), "sha256:"))

	err := os.MkdirAll(blobPath, 0755)
	if err != nil {
		return blobPath, fmt.Errorf("create folder failed: %v", err)
	}

	// blob 文件
	tarFilePath := filepath.Join(blobPath, saveProps.name)

	// 检查文件是否已存在
	if FileExists(tarFilePath) {
		Logger.Infof("Blob already exists, skipping: %s\n", desc.Digest)
		return blobPath, nil
	}

	// 创建目标文件
	tarFile, err := os.Create(tarFilePath)
	if err != nil {
		return blobPath, fmt.Errorf("failed to create blob file: %v", err)
	}
	defer tarFile.Close()

	// 获取 blob 读取器
	blobReader, size, err := d.src.GetBlob(d.ctx, types.BlobInfo{
		Digest: desc.Digest,
		Size:   desc.Size,
	}, none.NoCache)

	if err != nil {
		return blobPath, fmt.Errorf("failed to get blob reader: %v", err)
	}
	defer blobReader.Close()

	// 复制 blob 内容
	copied, err := io.Copy(tarFile, blobReader)
	if err != nil {
		return blobPath, fmt.Errorf("failed to copy blob content: %v", err)
	}

	// 验证大小
	if copied != size {
		return blobPath, fmt.Errorf("blob size mismatch: expected %d, got %d", size, copied)
	}

	Logger.Infof("  Successfully downloaded blob: %s (%d bytes)\n", desc.Digest, copied)
	return blobPath, nil
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
