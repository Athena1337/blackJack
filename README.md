# blackJack

`blackJack`是由[Lumos框架](https://github.com/Athena1337/Lumos)的核心侦查功能独立出来的小工具

用于从大量的资产中进行Web指纹探测，提取`有用`的系统

## Usage

### help

```bash
λ blackJack.exe -h

██████╗ ██╗      █████╗  ██████╗██╗  ██╗     ██╗ █████╗  ██████╗██╗  ██╗
██╔══██╗██║     ██╔══██╗██╔════╝██║ ██╔╝     ██║██╔══██╗██╔════╝██║ ██╔╝
██████╔╝██║     ███████║██║     █████╔╝      ██║███████║██║     █████╔╝
██╔══██╗██║     ██╔══██║██║     ██╔═██╗ ██   ██║██╔══██║██║     ██╔═██╗
██████╔╝███████╗██║  ██║╚██████╗██║  ██╗╚█████╔╝██║  ██║╚██████╗██║  ██╗
╚═════╝ ╚══════╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝ ╚════╝ ╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝ v1.0.0

Usage of blackJack.exe:
  -d    enable debug mode
  -i string
        Analyse target favicon fingerprint
  -l string
        the list file contain mutilple target url
  -o string
        output file
  -p string
        http proxy ,Ex: http://127.0.0.1:8080
  -t int
        request thread, default 50 (default 50)
  -time int
        request timeout (default 5)
  -u string
        single target url
```

### Running with file input

```bash
λ blackJack -l urls.txt
```

### Running with single url

```bash
λ blackJack -u https://google.com
```

## Features

+ 自动协议识别
+ WAF、CDN识别
+ 指纹覆盖优化，避免302跳转、CDN、均衡负载导致识别失效
+ 集成`icon hash`生成
+ 新增指纹至748条

## Thanks

探测功能的灵感和基本指纹库来自[EHole](https://github.com/EdgeSecurityTeam/EHole)

并发与一些细节参考了[httpx](https://github.com/projectdiscovery/httpx)

为了兼顾准确率，并发效率上，比`ehole`低，与`httpx`相差无几