# mini-container

一个简单的、迷你的容器技术探索项目（1600+行Go代码）

 **参考和致谢（含视频讲解）**：https://github.com/HobbyBear/tinydocker

# 架构 / 目录层级
![structure.png](assets%2Fstructure.png)

# 核心知识点
1. linux namespace 命名空间
2. linux union fs 联合文件系统 
3. linux cgroup 资源控制组
4. linux 文件根目录隔离
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


# 常见问题
1. Q：如何使得容器支持域名解析？
   
    A：在容器中配置`/etc/resolv.conf`文件，增加：
    ```bash
      nameserver 8.8.8.8
      nameserver 8.8.4.4
    ```



# Next

1. 引入cobra，增加更多可选命令（主要为增加cgroup参数，目前cgroup功能还未使用）
2. 增加：宿主机端口映射功能
3. 增加：用户目录挂载功能
4. 增加：后台运行功能
