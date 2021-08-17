## jiacrontab

[![Build Status](https://travis-ci.org/iwannay/jiacrontab.svg?branch=dev)](https://travis-ci.org/iwannay/jiacrontab) 

简单可信赖的任务管理工具

### v2.0.0版发布


### [❤jiacrontab 最新版下载点这里❤ ](https://jiacrontab.iwannay.cn/download/)

    1.自定义job执行  
    2.允许设置job的最大并发数  
    3.每个脚本都可在web界面下灵活配置，如测试脚本运行，查看日志，强杀进程，停止定时...  
    4.允许添加脚本依赖（支持跨服务器），依赖脚本提供同步和异步的执行模式  
    5.支持异常通知  
    6.支持守护脚本进程  
    7.支持节点分组


### 架构

<img src="https://raw.githubusercontent.com/iwannay/static_dir/master/jiacrontab_arch.png" width="50%"/>

### 说明

jiacrontab 由 jiacrontab_admin，jiacrontabd 两部分构成，两者完全独立通过 rpc 通信  
jiacrontab_admin：管理后台向用户提供web操作界面  
jiacrontabd：负责job数据存储，任务调度  


### 安装

#### 二进制安装

1.[下载](https://jiacrontab.iwannay.cn/download/) 二进制文件。

2.解压缩进入目录(jiarontab_admin,jiacrontabd)。

3.运行

```sh
$ nohup ./jiacrontab_admin &> jiacrontab_admin.log &
$ nohup ./jiacrontabd &> jiacrontabd.log &

## 建议使用systemd守护
```

#### v2.x.x源码安装

1.安装 git，golang(version 1.12.x)；可参考官网。
2.安装运行

```sh
$ git clone git@github.com:iwannay/jiacrontab.git
$ cd jiacrontab
# 配置代理
$ go env -w GONOPROXY=\*\*.baidu.com\*\*              ## 配置GONOPROXY环境变量,所有百度内代码,不走代理
$ go env -w GONOSUMDB=\*                              ## 配置GONOSUMDB,暂不支持sumdb索引
$ go env -w GOPROXY=https://goproxy.baidu.com         ## 配置GOPROXY,可以下载墙外代码

# 编译
# 注意需要先编译前端(https://github.com/jiacrontab/jiacrontab-frontend)
# 再安装go-bindata
# 然后assets指定前端资源编译后的位置
$ GO111MODULE=on go get -u github.com/go-bindata/go-bindata/v3/go-bindata
$ make build assets=$HOME/project/jiacrontab-frontend/build

$ cd build/jiacrontab/jiacrontab_admin/
$ nohup ./jiacrontab_admin &> jiacrontab_admin.log &

$ cd build/jiacrontab/jiacrontabd/
$ nohup ./jiacrontabd &> jiacrontabd.log &
```

<font color="red" size="3">浏览器访问 host:port (eg: localhost:20000) 即可访问管理后台</font>

#### docker 安装
```sh
# 下载镜像
$ docker pull iwannay/jiacrontab:2.3.0

# 创建自定义网络
$ docker network create mybridge

# 启动jiacrontab_admin
# 需要指定配置文件目录时需要先挂载目录，然后-config指定
$  docker run --network mybridge --name jiacrontab_admin -p 20000:20000 -it iwannay/jiacrontab:2.3.0 ./jiacrontab_admin

# 启动jiacrontabd
# 需要指定配置文件目录时需要先挂载目录，然后-config指定
docker run -v $(pwd)/jiacrontabd:/config --name jiacrontabd --network mybridge -it iwannay/jiacrontab:2.3.0 ./jiacrontabd -config /config/jiacrontabd.ini

```

### 升级版本

1、下载新版本压缩包，并解压。

2、替换旧版jiacrontab_admin,jiacrontabd为新版执行文件

3、运行

### 基本使用

#### 定时任务

##### 超时设置和超时操作

超时后会进行设置的超时操作 默认值为 0 不判断超时

##### 最大并发数

最大并发数控制同一job同一个时刻最多允许存在的进程数，默认最大并发数为1，当前一次未执行结束时则放弃后续执行。    
防止脚本无法正常退出而导致系统资源耗尽

##### 添加依赖

依赖就是用户脚本执行前，需要先执行依赖脚本，只有依赖脚本执行完毕才会执行当前脚本。  
1. **并发执行**  
并发执行依赖脚本，任意一个脚本出错或超时不会影响其他依赖脚本，但是会中断用户job
2. **同步执行**  
同步执行依赖脚本，执行顺序为添加顺序，如果有一个依赖脚本出错或超时，则会中断后继依赖，以及用户job

##### 脚本异常退出通知

如果脚本退出码不为0，则认为是异常退出

#### 常驻任务

常驻任务检查脚本进程是否退出，如果退出再次重启，保证脚本不停运行。  
注意：不支持后台进程。

## 附录

### 错误日志

错误日志存放在配置文件设置的目录下  
定时任务为 logs/crontab_task  
定时任务为 logs/daemon_task
日志文件准确为日期目录下的 ID.log （eg: logs/crontab_task/2018/01/01/1.log）

#### 错误日志信息

1. 正常错误日志  
程序原因产生的错误日志
2. 自定义错误日志  
程序中自定义输出的信息，需要在输出信息后面加入换行

### 截图
<img src="https://raw.githubusercontent.com/iwannay/static_dir/master/jiacrontab_preview_2.0.0_1.png" width="50%"/><img src="https://raw.githubusercontent.com/iwannay/static_dir/master/jiacrontab_preview_2.0.0_2.png" width="50%" />

### 演示地址

[2.0.0版本演示地址](http://jiacrontab-spa.iwannay.cn/) 账号：test 密码：123456

### 商务合作
<img src="https://raw.githubusercontent.com/iwannay/static_dir/master/1555767089.jpg" width="250" alt="合作"/>


### 赞助
本项目花费了作者大量时间，如果你觉的该项目对你有用，或者你希望该项目有更好的发展，欢迎赞助。
<img src="https://raw.githubusercontent.com/iwannay/jiacrontab/master/admire.jpg" width="250" alt="赞助"/><img src="https://raw.githubusercontent.com/iwannay/jiacrontab/master/qq.png" width="250" alt="qq群"/>
