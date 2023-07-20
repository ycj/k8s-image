# 镜像下载转发工具

# 用法

1. 下载镜像

    拉取非docker.io仓库的镜像，docker.io镜像不建议使用此工具下载    

    ```bash
    k8s-image pull registry.k8s.io/kube-apiserver:v1.18.0
    ```

2. 上传镜像

    上传本地镜像到指定仓库

    ```bash
    k8s-image push registry.k8s.io/kube-apiserver:v1.18.0  xxx.xxx.xxx.xxx:xxxx
    ```

3. 转发镜像

    将远程仓库镜像转发到指定仓库，可以简单理解是pull+push的组合
    
    ```bash
    k8s-image redirect registry.k8s.io/kube-apiserver:v1.18.0 xxx.xxx.xxx.xxx:xxxx
    ```
