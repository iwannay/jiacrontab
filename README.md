# jiacrontab

[![Build Status](https://travis-ci.org/iwannay/jiacrontab.svg?branch=dev)](https://travis-ci.org/iwannay/jiacrontab) 

简单可信赖的任务管理工具

## 2.0.0-alpha 测试版发布(正处于测试阶段，不建议生产使用。)
生产建议使用1.4.x版本

## [🔴jiacrontab 最新版下载点这里 ](https://jiacrontab.iwannay.cn/download/)

1.允许设置每个脚本的超时时间，超时操作可选择邮件通知管理者，或强杀脚本进程。  
2.允许设置脚本的最大并发数。  
3.一台 server 管理多个 client。  
4.每个脚本都可在 server 端灵活配置，如测试脚本运行，查看日志，强杀进程，停止定时...。  
5.允许添加脚本依赖（支持跨服务器），依赖脚本提供同步和异步的执行模式。  
6.友好的 web 界面，方便用户操作。  
7.脚本出错时可选择邮箱通知多人。  
8.支持常驻任务,任务失败后可配置自动重启。  
9.支持管道操作。

## 结构

![alt 架构](https://raw.githubusercontent.com/iwannay/static_dir/master/jiacrontab_arch.PNG)

## 说明

jiacrontab 由 server，client 两部分构成，两者完全独立通过 rpc 通信。  
server：向用户提供可视化界面，调度多个 client。  
client：实现定时逻辑，隔离用户脚本，将 client 布置于多台服务器上可由 server 统一管理。
每个脚本的定时格式完全兼容 linux 本身的 crontab 脚本配置格式。

## 安装

#### 二进制安装

1.[下载](https://jiacrontab.iwannay.cn/download/) 二进制文件。

2.解压缩进入目录(server,client)。

3.运行

```sh
$ nohup ./jiaserver &> server.log &
$ nohup ./jiaclient &> client.log &
```

### 1.4.\*源码安装

1.安装 git，golang(version 1.11.x)；可参考官网。  
2.安装运行

```sh
$ cd $GOPATH/src
$ git clone git@github.com:iwannay/jiacrontab.git
$ cd jiacrontab
$ make build

$ cd app/jiacrontab/server
$ nohup ./jiaserver &> jiaserver.log &

$ cd app/jiacrontab/client
$ nohup ./jiaclient &> jiaclient.log &
```

<font color="red" size="3">浏览器访问 host:port (eg: localhost:20000) 即可访问可视化界面</font>

### 升级至 1.4.x

1、下载新版本压缩包，并解压。

2、如果旧版存在 server/.data 和 client/.data 则拷贝至新版相同位置

3、拷贝 server/data、client/data、server/server.ini、client/client.ini 至新版相同位置

4、运行新版

## 配置文件

### 服务端配置文件 server.ini

```
;允许使用的command  可以在后面添加自己的command,用逗号隔开
allow_commands = php,/usr/local/bin/php,python,node,curl,wget,sh,uname,grep,cat,sleep
```

### 客户端配置文件 client.ini

```
; pprof 监听地址
pprof_addr = :20002

; 本机rpc监听地址
listen= :20001

; 推送给server的地址 host:port 在可视化界面展示
; 写本机IP推送给server之后 server记录下这个ip, server发送请求通过此地址
local_addr = localhost:20001

; server 地址 服务器 host:port 除非在同一台机器部署双端 否则需要更改
server_addr =localhost:20003

; 日志目录
dir = logs
; 自动清理大于一个月或者单文件体积大于1G的日志文件
clean_task_log = true
```

## 基本使用

### 定时任务

1. 超时设置和超时操作  
   超时后会进行设置的超时操作 默认值为 0 不判断超时

2. 最大并发数  
   最大并发数 控制 同时有几个脚本进程  
   默认最大并发数为 1，若不设置超时时间，当定时任务第二次执行时，若上一次执行还未完成  
   则会 kill 上一个脚本，进行本次执行。  
   防止脚本无法正常退出而导致系统资源耗尽

3. 添加依赖  
   依赖就是用户脚本执行前，需要先执行依赖脚本，只有依赖脚本执行完毕才会执行当前脚本。  
   3.1 并发执行  
   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;并发执行依赖脚本，任意一个脚本出错或超时不会影响其他依赖脚本，但是会中断用户脚本  
   3.2 同步执行  
   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;同步执行依赖脚本，执行顺序为添加顺序，如果有一个依赖脚本出错或超时，则会中断后继依赖，以及用户脚本

4. 脚本异常退出通知
   如果脚本退出码不为 0，则认为是异常退出

### 常驻任务

常驻任务检查脚本进程是否退出，如果退出再次重启，保证脚本不停运行  
其他同 定时任务

## 附录

### 错误日志

错误日志存放在配置文件设置的目录下  
定时任务为 logs/crontab_task  
定时任务为 daemon_task
日志文件准确为日期目录下的 ID.log （eg: logs/crontab_task/2018/01/01/1.log）

#### 错误日志信息

1. 正常错误日志  
   程序原因产生的错误日志
2. 自定义错误日志  
   程序中自定义输出 需要在输出信息后面加入换行 （eg: `echo ‘自定义错误信息’.“\n”`）

## 1.4.\*截图

![alt 截图1](https://raw.githubusercontent.com/iwannay/static_dir/master/jiacrontab_preview_1.4.0_list.png)

![alt 截图2](https://raw.githubusercontent.com/iwannay/static_dir/master/jiacrontab_preview_1.4.0_edit.png)

## 演示地址

[1.4.\*版本演示地址](http://jiacrontab.iwannay.cn/) 账号：admin 密码：123456

## QQ群号：813377930
<img src="https://github.com/iwannay/jiacrontab/blob/dev/qq.png" width="250" alt="qq群"/>

## 赞助
本项目花费了作者大量时间，如果你觉的该项目对你有用，或者你希望该项目有更好的发展，欢迎赞助。
<img src="https://github.com/iwannay/jiacrontab/blob/dev/alipay.jpg" width="250" alt="赞助"/>
