# jiacrontab
提供可视化界面的定时任务管理工具。

1.允许设置每个脚本的超时时间，超时操作可选择邮件通知管理者，或强杀脚本进程。  
2.允许设置脚本的最大并发数。  
3.可一台server管理多个client。  
4.每个脚本都可在server端灵活配置，如测试脚本运行，查看日志，强杀进程，停止定时...。

## 说明
jiacrontab由server，client两部分构成，两者完全独立通过rpc通信。  
server：向用户提供可视化界面，调度多个client。  
client：实现定时逻辑，隔离用户脚本，将client布置于多台服务器上可由server统一管理。
每个脚本的定时格式完全兼容linux本身的crontab脚本配置格式。

## 安装
#### 二进制安装  
1.[下载](http://git.wzjg520.com/wzjg520/jiacrontab/releases) 二进制文件。  

2.解压缩进入目录。  

3.运行  
```sh
$ nohub ./server &> server.log &
$ nohub ./client &> client.log &     
```
### 源码安装
1.安装git，golang；可参考官网。  
2.安装运行
```sh
$ cd $GOPATH/src
$ git clone http://git.wzjg520.com:/wzjg520/jiacrontab.git 
$ go get -u github.com/dgrijalva/jwt-go
$ go get -u gopkg.in/ini.v1

$ cd $GOPATH/src/jiacrontab/server
$ go build .
$ nohub ./server &> server.log &

$ cd $GOPATH/src/jiacrontab/task
$ go build .
$ nohub ./client &> client.log & 
``` 

### 截图
![alt 截图](http://static.wzjg520.com/view?id=8-1496904294.jpg)  

![alt 截图2](http://static.wzjg520.com/view?id=8-1496904302.jpg)

### 演示地址
[demo](http://182.92.223.12:20000) 账号：admin 密码：123456
