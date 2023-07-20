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
1. k8s-image pull     gcr.io/google_containers/kube-apiserver-amd64:v1.9.0
2. k8s-image push     gcr.io/google_containers/kube-apiserver-amd64:v1.9.0  <ip:port>
3. k8s-image redirect gcr.io/google_containers/kube-apiserver-amd64:v1.9.0  <ip:port>
*/

func main() {
	fmt.Println(os.Args)

	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Println("v0.2.0")
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
		newImageUri := imageUriConvertToDockerHub(originImageUri)
		fmt.Println("拉取镜像:", newImageUri)
		imagePull(newImageUri)

		if originImageUri != newImageUri {
			fmt.Println("重命名标签:", originImageUri)
			renameTag(newImageUri, originImageUri)
			fmt.Println("删除标签:", newImageUri)
			deleteTag(newImageUri)
		}
	case "push":
		originImageUri := os.Args[2]
		address := os.Args[3]
		newImageUri := imageUriConvertToPrivateRegistry(originImageUri, address)
		imagePush(newImageUri)
		if originImageUri != newImageUri {
			deleteTag(newImageUri)
		}
	case "redirect":
		originImageUri := os.Args[2]
		address := os.Args[3]
		newImageUri := imageUriConvertToDockerHub(originImageUri)
		fmt.Println("拉取镜像:", newImageUri)
		imagePull(newImageUri)

		if originImageUri != newImageUri {
			fmt.Println("重命名标签:", originImageUri)
			renameTag(newImageUri, originImageUri)
			fmt.Println("删除标签:", newImageUri)
			deleteTag(newImageUri)
		}

		newImageUri = imageUriConvertToPrivateRegistry(originImageUri, address)
		echo("匹配私有仓库地址：", newImageUri)
		imagePush(newImageUri)
		if originImageUri != newImageUri {
			echo("删除标签：", newImageUri)
			deleteTag(newImageUri)
		}
		echo("删除标签：", originImageUri)
		deleteTag(originImageUri)
	default:
		fmt.Println("不支持的操作")
	}
}

// 任意类型打印输出
func echo(args ...any) {
	fmt.Println(args...)
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

// 转换镜像地址匹配私有仓库
func imageUriConvertToPrivateRegistry(imageUri string, address string) string {
	data := strings.Split(imageUri, "/")
	dataLen := len(data)
	registry := data[0]

	if dataLen == 4 && registry == "ghcr.io" {
		return fmt.Sprintf("%s/%s/%s/%s", address, data[1], data[2], data[3])
	}

	if dataLen == 3 && (registry == "gcr.io" || registry == "quay.io" || registry == "registry.k8s.io" || registry == "k8s.gcr.io" || registry == "ghcr.io") {
		return fmt.Sprintf("%s/%s/%s", address, data[1], data[2])
	}

	if dataLen == 2 && (registry == "k8s.gcr.io" || registry == "registry.k8s.io") {
		return fmt.Sprintf("%s/%s", address, data[1])
	}

	return imageUri
}

// 转换镜像地址
func imageUriConvertToDockerHub(imageUri string) string {
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
func imagePush(imageUri string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	out, err := cli.ImagePush(context.Background(), imageUri, types.ImagePushOptions{})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)
}
