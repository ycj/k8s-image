package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

/*
usage:
1. k8s-image pull gcr.io/google_containers/kube-apiserver-amd64:v1.9.0
2. k8s-image push gcr.io/google_containers/kube-apiserver-amd64:v1.9.0  <ip:port>
*/

func main() {
	// 1. pull image
	// 2. push image
	fmt.Println(os.Args)

	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println("v0.1.0")
		return
	} else if len(os.Args) < 3 {
		fmt.Println(`
		使用说明: 
		  k8s-image pull     <image>           从公共仓库拉取镜像
		  k8s-image push     <image> <ip:port> 推送镜像到私有仓库
		  k8s-image redirect <image> <ip:port> 从公共仓库拉取镜像推送到私有镜像
		`)
		return
	}

	switch os.Args[1] {
	case "pull":
		originImageUri := os.Args[2]
		newImageUri := imageUriTransfer(originImageUri)
		fmt.Println("拉取镜像:", newImageUri)
		imagePull(newImageUri)
		fmt.Println("重命名标签:", originImageUri)
		renameTag(newImageUri, originImageUri)
		fmt.Println("删除标签:", newImageUri)
		deleteTag(newImageUri)
	case "push":
		imagePush(os.Args[2], os.Args[3])
	default:
		fmt.Println("不支持的操作")
	}
}

// 删除镜像
func deleteTag(imageUri string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	cli.ImageRemove(context.Background(), imageUri, types.ImageRemoveOptions{})
}

// 重命名镜像标签
func renameTag(originImageUri, newImageUri string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	err = cli.ImageTag(context.Background(), originImageUri, newImageUri)
	if err != nil {
		panic(err)
	}
}

// 转换镜像地址
// gcr.io/google_containers/kube-apiserver-amd64:v1.9.0  -> anjia0532/google_containers.kube-apiserver-amd64:v1.9.0
// registry.k8s.io/pause:1.0 -> anjia0532/google_containers.pause:1.0
// k8s.gcr.io/pause-amd64:3.1 -> anjia0532/google_containers.pause-amd64:3.1
func imageUriTransfer(imageUri string) string {
	data := strings.Split(imageUri, "/")
	dataLen := len(data)
	registry := data[0]

	if dataLen == 4 && registry == "ghcr.io" {
		return fmt.Sprintf("anjia0532/ghcr.%s.%s.%s", data[1], data[2], data[3])
	}

	if dataLen == 3 && registry == "gcr.io" {
		return fmt.Sprintf("anjia0532/%s.%s", data[1], data[2])
	}

	if dataLen == 3 && (registry == "registry.k8s.io" || registry == "k8s.gcr.io") {
		return fmt.Sprintf("anjia0532/google-containers.%s.%s", data[1], data[2])
	}

	if dataLen == 2 && (registry == "k8s.gcr.io" || registry == "registry.k8s.io") {
		return fmt.Sprintf("anjia0532/google_containers.%s", data[1])
	}

	return imageUri
}

// 拉取镜像
func imagePull(imageUri string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	out, err := cli.ImagePull(context.Background(), imageUri, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)
}

// 推送镜像
func imagePush(imageUri, address string) {

}
