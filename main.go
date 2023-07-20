package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
		fmt.Println("v0.2.4")
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
		imagePull(newImageUri)

		if originImageUri != newImageUri {
			fmt.Println("还原标签:", originImageUri)
			renameTag(newImageUri, originImageUri)
			fmt.Println("删除临时标签:", newImageUri)
			deleteTag(newImageUri)
		}
	case "push":
		originImageUri := os.Args[2]
		address := os.Args[3]
		newImageUri := imageUriConvertToPrivateRegistry(originImageUri, address)
		renameTag(originImageUri, newImageUri)
		imagePush(newImageUri)
		echo("删除临时标签：", newImageUri)
		deleteTag(newImageUri)

	case "redirect":
		originImageUri := os.Args[2]
		address := os.Args[3]
		newImageUri := imageUriConvertToDockerHub(originImageUri)
		imagePull(newImageUri)
		if originImageUri != newImageUri {
			fmt.Println("还原标签:", originImageUri)
			renameTag(newImageUri, originImageUri)
			fmt.Println("删除标签:", newImageUri)
			deleteTag(newImageUri)
		}

		privateImageUri := imageUriConvertToPrivateRegistry(originImageUri, address)
		echo("私有标签：", privateImageUri)
		renameTag(originImageUri, privateImageUri)
		imagePush(privateImageUri)
		echo("删除私有标签：", privateImageUri)
		deleteTag(privateImageUri)
		echo("删除临时镜像：", originImageUri)
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

	if dataLen == 2 {
		if registry == "k8s.gcr.io" || registry == "registry.k8s.io" { // example: k8s.gcr.io/kube-apiserver:v1.9.0
			return fmt.Sprintf("%s/%s", address, data[1])
		} else { // example: apache/flink:1.11.2-scala_2.12-java11
			return fmt.Sprintf("%s/%s/%s", address, data[0], data[1])
		}
	}

	if dataLen == 1 { //example: nginx:1.25.1
		return fmt.Sprintf("%s/%s", address, data[0])
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
	echo("拉取镜像：", imageUri)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	out, err := cli.ImagePull(context.Background(), imageUri, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	// io.Copy(os.Stdout, out)
	decoder := json.NewDecoder(out)
	for {
		var s Status
		if err := decoder.Decode(&s); err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		if s.Error != "" {
			panic(s.ErrorDetail.Message)
		}
		if s.Progress != nil {
			if s.Progress.Current < s.Progress.Total {
				fmt.Printf("%20s %s %d/%d\r", s.Status, s.Id, s.Progress.Current, s.Progress.Total)
			} else {
				fmt.Printf("%20s %s %d/%d\n", s.Status, s.Id, s.Progress.Current, s.Progress.Total)
			}
		} else {
			fmt.Println(s.Status)
		}
	}
}

// 推送镜像
func imagePush(imageUri string) {
	echo("推送镜像：", imageUri)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	username := os.Getenv("DOCKER_USERNAME")
	password := os.Getenv("DOCKER_PASSWORD")
	if username == "" || password == "" {
		panic("请设置环境变量：DOCKER_USERNAME、DOCKER_PASSWORD")
	}
	authConfig := types.AuthConfig{
		Username: username,
		Password: password,
	}
	jsonData, err := json.Marshal(authConfig)
	if err != nil {
		panic(err)
	}
	authStr := base64.URLEncoding.EncodeToString(jsonData)
	out, err := cli.ImagePush(context.Background(), imageUri, types.ImagePushOptions{
		All:           false,
		RegistryAuth:  authStr,
		PrivilegeFunc: nil,
	})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)
}

// 解析Docker API的响应
type Progress struct {
	Current int64 `json:"current"`
	Total   int64 `json:"total"`
}
type Status struct {
	Status      string    `json:"status"`
	Progress    *Progress `json:"progressDetail"`
	Id          string    `json:"id"`
	Error       string    `json:"error,omitempty"`
	ErrorDetail struct {
		Message string `json:"message"`
	} `json:"errorDetail,omitempty"`
}
