# mini-container

一个简单的、迷你的容器技术探索项目

 **参考和致谢（含视频讲解）**：https://github.com/HobbyBear/tinydocker

# 核心知识点
1. linux namespace 命名空间
2. linux union fs 联合文件系统 
3. linux cgroup 资源控制组
4. linux chroot 文件根目录隔离
5. linux network知识，包括veth、bridge等



# 涉及到的技术栈
1. go 1.21
2. linux 相关系统调用

# 使用
## 前提：
1. 机器环境最好为ubuntu 22.04（开发所使用的环境）
2. 宿主机需执行：
    ```bash
    $ iptables -A FORWARD -j ACCEPT 
    $ echo 1 > /proc/sys/net/ipv4/ip_forward
    ```

## 目前支持以下命令：

1. ./mini-container run [container name] [image path] [entry point] [args...]
    
    For example: ./mini-container run test1 / /bin/sh 

2. ./mini-container ls
3. ./mini-container rm [container name]
4. ./mini-container clear
5. ./mini-container start [container name]
6. ./mini-container stop [container name]


# Next

1. 宿主机端口映射
2. 引入cobra，增加更多可选命令（主要为增加cgroup参数，目前未使用）
