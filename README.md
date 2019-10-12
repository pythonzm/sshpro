# sshpro
批量管理工具，阉割版ansible

## 使用
1. git clone https://github.com/pythonzm/sshpro.git
2. cd sshpro && go build

当前目录下会生成可执行文件sshpro，将文件链接到bin下：`ln -sv $PWD/sshpro /bin/sshpro`

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
