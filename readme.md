xray的poc很多，但是xray又不开源，无法进行二次开发改造。
于是根据xray文档中的poc检测思路，用cel-go写了个轮子。自测目前应该是支持除了xray中的除反连之外的poc。

目前支持四种检测方式
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