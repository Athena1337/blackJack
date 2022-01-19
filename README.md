# blackJack

`blackJack`是由 [Lumos框架](https://github.com/Athena1337/Lumos) 的核心侦查功能独立出来的小工具

用于从大量的资产中进行Web指纹探测，提取`有用`的系统，并能对探测后的目标进行目录扫描和备份文件扫描

## Usage

### help

```bash
λ blackJack -h

██████╗ ██╗      █████╗  ██████╗██╗  ██╗     ██╗ █████╗  ██████╗██╗  ██╗
██╔══██╗██║     ██╔══██╗██╔════╝██║ ██╔╝     ██║██╔══██╗██╔════╝██║ ██╔╝
██████╔╝██║     ███████║██║     █████╔╝      ██║███████║██║     █████╔╝
██╔══██╗██║     ██╔══██║██║     ██╔═██╗ ██   ██║██╔══██║██║     ██╔═██╗
██████╔╝███████╗██║  ██║╚██████╗██║  ██╗╚█████╔╝██║  ██║╚██████╗██║  ██╗
╚═════╝ ╚══════╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝ ╚════╝ ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝ v1.0.0

NAME:
   blackJack - Usage Menu

USAGE:
   blackJack [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   -d            Enable debug mode (default: false)
   -u value      Single target url
   -l value      The list file contain mutilple target url
   -t value      Request thread (default: 50)
   --time value  Request timeout (default: 30s)
   -r value      Max Retry attempts (default: 5)
   -o value      Output file
   -p value      http proxy ,Ex: http://127.0.0.1:8080
   -i value      Analyse target favicon fingerprint
   -b            Enable DirBrute for analyse target (default: false)
   --help, -h    show help (default: false)
```

### Running with file input

```bash
λ blackJack -l urls.txt
```

### Running with single url

```bash
λ blackJack -u https://google.com
```

### Dir Brute

对指纹探测后的目标进行目录扫描和备份文件扫描

使用`simhash`算法识别页面，必须与伪404页面、主页不相似的页面才被打印

并能动态识别不同路径高相似页面然后进行过滤 ，实测0误报
```bash
λ blackJack -l urls.txt -b
```
字典默认使用[blackJack-Dicts](https://github.com/t43Wiu6/blackJack-Dicts)

首次使用时字典默认下载到当前用户的`home/.blackJack`目录, 具体分类请看该项目, 欢迎pr和自定义修改

由于终端打印重用覆盖会导致刷大量换行符，虽不影响使用但不建议用`win`，推荐`Linux`

### More
```bash
λ blackJack -l urls.txt -d # 开启Debug模式
λ blackJack -i google.com  # 检测提供域名的favicon hash，用于fofa等测绘引擎
```

## Update Logs

### V1.0
+ 新增目录扫描和备份文件扫描功能

### V1.0-beta3
+ 更换底层HTTP请求库
+ 重构并发模块，减少开销
+ 提高稳定性
+ 速度提升500%

### V1.0-beta2
+ 深度去重指纹
+ 重构指纹结构
+ 重构优化部分处理模块
+ 更新指纹1833条，合计2581条

### v1.0-beta1 
+ 自动协议识别
+ WAF、CDN识别
+ 指纹覆盖优化，避免302跳转、CDN、均衡负载导致识别失效
+ 集成`icon hash`生成
+ 集成指纹748条

## Thanks

探测功能的灵感和原基本指纹库来自[EHole](https://github.com/EdgeSecurityTeam/EHole)
一些细节参考了[httpx](https://github.com/projectdiscovery/httpx)