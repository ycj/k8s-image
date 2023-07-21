# 镜像下载转发工具

# 用法

1. 下载镜像

    拉取非docker.io仓库的镜像，docker.io镜像不建议使用此工具下载    

    ```bash
    k8s-image pull registry.k8s.io/kube-apiserver:v1.18.0
    ```

    如果提示拉去的镜像提示不存在，请在此地址 https://github.com/anjia0532/gcr.io_mirror/issues 填写ISSUE触发镜像拉取成功后再次执行  k8s-image pull [你的镜像]

    如果遇到镜像TAG匹配失败，请在此项目提交issue反馈

3. 上传镜像

    上传本地镜像到指定仓库

    ```bash
    k8s-image push registry.k8s.io/kube-apiserver:v1.18.0  xxx.xxx.xxx.xxx:xxxx
    ```

4. 转发镜像

    将远程仓库镜像转发到指定仓库，可以简单理解是pull+push的组合
    
    ```bash
    k8s-image redirect registry.k8s.io/kube-apiserver:v1.18.0 xxx.xxx.xxx.xxx:xxxx
    ```
