package main

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/containers/image/v5/types"
)

type TarInfo struct {
	Ref          types.ImageReference
	ConfigDigest string
	LayersDigest []string
	ImageInfo    DockerImageV2

	folderPath string
}

func (t *TarInfo) BuildTar() {

	err := t.mkdirTmp()
	if err != nil {
		log.Fatalf("Failed to create tmp directory: %v", err)

		return
	}

	// 根目录生成 xx.json
	err = t.buildConfigjson()
	if err != nil {
		log.Fatalf("Failed to build config.json: %v", err)

		return
	}

	err = t.buildLayers()
	if err != nil {
		log.Fatalf("Failed to build layers: %v", err)

		return
	}

	err = t.buildRepositoriesjson()
	if err != nil {
		log.Fatalf("Failed to build repositories.json: %v", err)

		return
	}

	// 根目录生成 manifest.json
	err = t.buildManifestjson()
	if err != nil {
		log.Fatalf("Failed to build manifest.json: %v", err)

		return
	}

	err = t.packTar()
	if err != nil {
		log.Fatalf("Failed to pack tar: %v", err)

		return
	}
}

func (t *TarInfo) packTar() error {

	tarFilePath := t.buildTarName()

	// 检查文件是否已存在
	if FileExists(tarFilePath) {
		fmt.Printf("Tar already exists, skipping: %s\n", tarFilePath)
		return nil
	}
	err := CreateTar(t.folderPath, tarFilePath)
	if err != nil {
		return fmt.Errorf("failed to create tar file: %v", err)
	}
	fmt.Printf("  Successfully created tar: %s\n", tarFilePath)
	return nil
}

func (t *TarInfo) buildTarName() string {
	return filepath.Join("output", t.Ref.DockerReference().Name()+"_"+t.ConfigDigest+".tar")
}

func (t *TarInfo) mkdirTmp() error {
	// 创建 tmp 文件夹
	tmpPath := filepath.Join("tmp", t.ConfigDigest)

	err := os.MkdirAll(tmpPath, 0755)
	if err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	t.folderPath = tmpPath

	return nil
}

func (t *TarInfo) buildLayers() error {
	for _, layerDigest := range t.LayersDigest {
		oriLayerTar := filepath.Join("cache", "layers", layerDigest, "layer.tar")
		destLayerDir := filepath.Join(t.folderPath, layerDigest)
		err := os.MkdirAll(destLayerDir, 0755)
		if err != nil {
			return fmt.Errorf("创建目录失败: %v", err)
		}
		destLayerTar := filepath.Join(destLayerDir, "layer.tar")

		// copy file
		err = CopyFile(oriLayerTar, destLayerTar)
		if err != nil {
			return err
		}

	}
	return nil
}

// 根目录生成 xx.json
func (t *TarInfo) buildConfigjson() error {

	oriConfigJson := filepath.Join("cache", "config", t.ConfigDigest, "config.json")
	destConfigJson := filepath.Join(t.folderPath, t.ConfigDigest+".json")

	// copy file
	err := CopyFile(oriConfigJson, destConfigJson)
	if err != nil {
		return err
	}

	return nil
}

type Schema2Manifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

// 根目录生成 manifest.json
func (t *TarInfo) buildManifestjson() error {

	filePath := filepath.Join(t.folderPath, "manifest.json")

	// 检查文件是否已存在
	if FileExists(filePath) {
		fmt.Printf("file already exists, skipping: %s\n", filePath)
		return nil
	}

	// 创建目标文件
	targetFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create blob file: %v", err)
	}
	defer targetFile.Close()

	mainfest := Schema2Manifest{
		Config:   t.ConfigDigest + ".json",
		RepoTags: []string{t.Ref.DockerReference().Name() + ":latest"},
		Layers: func() []string {
			var layers []string
			for _, layerDigest := range t.LayersDigest {
				layers = append(layers, layerDigest+"/layer.tar")
			}
			return layers
		}(),
	}

	listData := make([]Schema2Manifest, 0)
	listData = append(listData, mainfest)

	return WriteJsonFile(targetFile, listData)
}

func (t *TarInfo) buildRepositoriesjson() error {
	// 创建  文件
	repoFilePath := filepath.Join(t.folderPath, "repositories")

	// 检查文件是否已存在
	if FileExists(repoFilePath) {
		fmt.Printf("Blob already exists, skipping: %s\n", repoFilePath)
		return nil
	}

	// 创建目标文件
	repoFile, err := os.Create(repoFilePath)
	if err != nil {
		return fmt.Errorf("failed to create blob file: %v", err)
	}
	defer repoFile.Close()

	// {
	//	"golang": {
	//		"1.24.2-alpine3.20": "d2c830c9c895b70b315e96d9b40aa6c5135ff03f44d8a3a447488c1e1661c062"
	//	}
	//}

	repoName := t.Ref.DockerReference().Name()
	tag := t.ImageInfo.Tag

	data := make(map[string]map[string]string)
	data[repoName] = map[string]string{
		tag: SlicesLast(t.LayersDigest),
	}

	return WriteJsonFile(repoFile, data)
}

func WriteJsonFile(repoFile *os.File, data any) error {

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	_, err = repoFile.Write(jsonData)
	if err != nil {
		return fmt.Errorf("failed to write JSON data: %v", err)
	}

	return nil
}

func CreateTar(srcDir, tarFilePath string) error {
	// 创建目标目录
	destDir := filepath.Dir(tarFilePath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// 创建tar文件
	tarFile, err := os.Create(tarFilePath)
	if err != nil {
		return fmt.Errorf("failed to create tar file: %v", err)
	}
	defer tarFile.Close()

	// 创建tar writer
	tw := tar.NewWriter(tarFile)
	defer tw.Close()

	// 遍历源目录并添加到tar
	err = filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 创建tar头
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// 保持目录结构正确
		header.Name, err = filepath.Rel(filepath.Dir(srcDir), file)
		if err != nil {
			return err
		}

		// 写入头部信息
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// 如果是文件而不是目录，则写入内容
		if !fi.IsDir() {
			file, err := os.Open(file)
			if err != nil {
				return err
			}
			defer file.Close()

			// 复制文件内容到tar
			_, err = io.Copy(tw, file)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create tar: %v", err)
	}

	return nil
}
