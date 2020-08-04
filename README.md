# sshpro
批量管理工具，阉割版ansible，支持批量执行命令以及批量复制文件和文件夹

主机与被管理机器无需安装任何依赖，也不用配置互信，直接使用即可

## 使用
1. git clone https://github.com/pythonzm/sshpro.git
2. 生成可执行文件

```bash
cd sshpro
go mod init
go build
ln -sv $PWD/sshpro /bin/sshpro  # 将可执行文件链接到/bin目录下，可选
```

3. （可选，如果需要使用配置文件则需要配置）将配置文件模板复制到$HOME下 `cp .sshpro.yaml $HOME`，然后根据自己的情况修改模板

### 查看使用帮助

```shell
[root@test sshpro]# sshpro -h
ansible的阉割版。。。

Usage:
  sshpro [flags]
  sshpro [command]

Available Commands:
  copy        传输本地文件到远程主机
  help        Help about any command

Flags:
  -c, --command string      执行命令
      --config string       指定配置文件 (default is $HOME/.sshpro.yaml)
      --go-num int          并发数 (default 5)
  -g, --group string        根据配置文件指定某个组执行命令,多个组用','隔开
  -h, --help                help for sshpro
      --host-net string     主机段，例如：192.168.1.0/24
      --host-range string   主机范围，例如：10.1.1.1-10.1.1.254
      --hosts string        远程主机IP, 可以是一个或多个，多个用主机用','隔开
  -k, --key string          主机密钥位置，使用绝对路径
  -P, --password string     远程用户密码
  -p, --port string         远程端口 (default "22")
  -u, --username string     远程使用用户 (default "root")
      --version             version for sshpro

Use "sshpro [command] --help" for more information about a command.
```

## 部分使用案例

### 批量执行命令（不使用配置文件）
```shell
[root@test sshpro]# sshpro --hosts 10.1.7.239,10.1.7.240 -u root -P 123456 -c uptime
10.1.7.239 | SUCCESS => 
 15:49:08 up 3 days,  5:52,  3 users,  load average: 0.00, 0.01, 0.05

10.1.7.240 | FAILED => dial tcp 10.1.7.240:22: i/o timeout
```

### 批量执行命令（使用配置文件，需要使用-g参数）
```shell
[root@test sshpro]# sshpro -g centos -c uptime    # centos组内主机全部执行uptime命令
10.1.7.239 | SUCCESS => 
 15:52:23 up 3 days,  5:56,  3 users,  load average: 0.00, 0.01, 0.05

10.1.7.198 | SUCCESS => 
 15:52:22 up 93 days,  3:49,  4 users,  load average: 0.05, 0.03, 0.00

10.1.7.199 | UNREACHABLE => dial tcp 10.1.7.199:22: connect: no route to host
```

### 批量复制文件（传递文件夹需要-r参数，不使用配置文件）
```
[root@test sshpro]# sshpro copy --hosts 10.1.7.239,10.1.7.238 -u root -P 123456 -s /tmp/aa/a.log -d /tmp/
10.1.7.239 | SUCCESS => 
/tmp/aa/a.log 传输完成

10.1.7.238 | UNREACHABLE => dial tcp 10.1.7.238:22: connect: no route to host
```

### 批量复制文件（传递文件夹需要-r参数，使用配置文件）
```shell
[root@test sshpro]# sshpro copy -g centos -s /tmp/aa/a.log -d /tmp/                                                 
10.1.7.198 | SUCCESS => 
/tmp/aa/a.log 传输完成

10.1.7.239 | SUCCESS => 
/tmp/aa/a.log 传输完成

10.1.7.199 | UNREACHABLE => dial tcp 10.1.7.199:22: connect: no route to host
```
