# mini-container

一个简单的、迷你的容器技术探索项目

# 核心知识点
1. linux namespace 命名空间
2. linux union fs 联合文件系统 
3. linux cgroup 资源控制组
4. linux chroot 文件根目录隔离
5. linux network知识，包括veth、bridge等

致谢和参考：https://github.com/HobbyBear/tinydocker

# 涉及到的技术栈
1. golang go1.21
2. linux namespace & cgroup & fs 

# 使用
## 前提：
1. 需要注意，请在linux环境下使用，且需要root权限
2. iptables -A FORWARD -j ACCEPT
3. echo 1 > /proc/sys/net/ipv4/ip_forward


## 目前支持以下命令：

1. ./mini-container run [container name] [image path] [entry point] [args...]
2. ./mini-container ls
3. ./mini-container rm [container name]
