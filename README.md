# jiacrontab
一款提供web可视化界面的定时任务管理工具.
## 说明
jiacrontab由server，client两部分构成，两者完全独立通过rpc通信。  
server：向用户提供可视化界面，调度多个client。  
client：实现定时逻辑，隔离用户脚本，将client布置于多台服务器上可由server统一管理。

## 安装
#### 二进制安装  
1.[下载](http://git.wzjg520.com/wzjg520/jiacrontab/releases) 二进制文件。  

2.解压缩进入目录。  

3.运行  
```sh
$ nohub ./jiacrontab_server 2>&1 > jiacrontab_server.log &
$ nohub ./jiacrontab_client 2>&1 > jiacontab_client.log &     
```
### 源码安装
1.安装git，golang；可参考官网。  
2.安装运行
```sh
$ cd $GOPATH/src
$ git clone http://www.wzjg520.com:9999/wzjg520/jiacrontab.git 
$ go get -u github.com/dgrijalva/jwt-go
$ go get -u gopkg.in/ini.v1

$ cd $GOPATH/src/jiacrontab
$ go build -o jiacrontab_server .
$ nohub ./jiacrontab_server 2>&1 > jiacrontab_server.log &

$ cd $GOPATH/src/jiacrontab/task
$ go build -o jiacrontab_client .
$ nohub ./jiacrontab_client 2>&1 > jiacontab_client.log & 
``` 

### 截图
![alt 截图](http://182.92.223.12:8080/view?id=15-1494839111.jpg)  

![alt 截图2](http://182.92.223.12:8080/view?id=15-1494839123.jpg)

### 演示地址
[demo](http://182.92.223.12:20000) 账号：admin 密码：123456
