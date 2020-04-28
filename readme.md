# Chinese
[xray](https://github.com/chaitin/xray) 提供了很多优秀简洁直观的POC，但是xray并不开源，无法进行二次开发改造。

于是根据xray文档中的检测poc的思路，用[cel-go](https://github.com/google/cel-go) 写了个轮子，方便批量检测。

目前应该是支持除反连之外的poc。

支持四种检测方式
一对一，单个目标执行单个poc
```
gopoc -t http://www.test.com -p poc.yaml
```
一对多，单个目标执行多个poc
```
gopoc -t http://www.test.com -P "poc/*"
```
多对一，多个目标执行单个poc
```
gopoc -l urls.txt -p poc.yaml
```
多对多，多个目标执行多个poc
```
gopoc -l urls.txt -P "pocs/*"
```

其他几个参数说明如下
```bash
-t 请求超时设置
-n 总并发数
-proxy 代理服务器，目前只测试了http代理
```

使用`-h`查看所有参数

# English
[xray](https://github.com/chaitin/xray) provides many excellent concise and intuitive POC, but xray is not open source and cannot be redeveloped.

So according to the idea of detecting poc in the chaitin xray document, I wrote a wheel with [cel-go](https://github.com/google/cel-go) to facilitate batch detection.

At present, it should support poc in addition to anti-link.
supports four detection methods

One-to-one, a single target performs a single poc
```
gopoc -t http://www.test.com -p poc.yaml
```
One-to-many, a single target performs multiple poc
```
gopoc -t http://www.test.com -P "poc / *"
```
Many to one, multiple targets execute a single poc
```
gopoc -l urls.txt -p poc.yaml
```
Many-to-many, multiple targets execute multiple poc
```
gopoc -l urls.txt -P "pocs / *"
```

Several other parameters are described below
```bash
-t request timeout setting
-n total number of concurrent
-proxy proxy server, currently only tested http proxy
```

Use `-h` to view all parameters