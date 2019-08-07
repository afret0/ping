# Usage

## config

```go
import  "goFrame/config"

conf := cong.GetConf()
fmt.Println(conf.app)
```

## build

### 打包配置文件

使用 packr2 打包 配置文件

`packr2 clean` 清除配置文件

### 测试环境一键部署

`./build_test.sh` 

# TODO

- [ ] 一键部署带个参数吧, 写死的好蠢
- [ ] 数据库链接
- [ ] 异常捕获
- [ ] 热更新  [fsnotiy](https://github.com/fsnotify/fsnotify/blob/master/example_test.go)  [fresh](https://github.com/gravityblast/fresh) 