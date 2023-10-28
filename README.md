# mini-container

一个简单的、迷你的容器技术探索项目

致谢和参考：https://github.com/HobbyBear/tinydocker

# 涉及到的技术栈
1. golang go1.21
2. linux namespace & cgroup & fs 

# 使用
需要注意，请在linux环境下使用，且需要root权限

目前支持以下命令：

1. ./mini-container run [container name] [image path] [entry point] [args...]
2. ./mini-container ls
3. ./mini-container rm [container name]
