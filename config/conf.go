package config

import (
	"bytes"
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"time"
)

type etc struct {
	App  string
	Port string
	Ping string
}

var conf *etc

func init() {
	initConfig()
	conf = new(etc)
	conf.Port = viper.GetString("port")
	conf.App = viper.GetString("app")
	conf.Ping = viper.GetString("ping")
}
func GetConf() *etc {
	return conf
}
func initConfig() {
	//读取文件 使用 packr 方法可以在 build 时 将配置文件打包
	box := packr.New("confBox", ".")
	configType := "yaml"
	defaultConfig, err := box.Find("app.yaml")
	if err != nil {
		panic(err)
	}
	//创建一个新实例  用来读取 app.yaml 初始文件 并保存为默认配置
	v := viper.New()
	v.SetConfigType(configType)
	err = v.ReadConfig(bytes.NewReader(defaultConfig))
	if err != nil {
		panic(err)
	}
	configs := v.AllSettings()
	for k, v := range configs {
		viper.SetDefault(k, v)
	}

	err = viper.BindEnv("ENV")
	if err != nil {
		panic(err)
	}
	env := viper.GetString("ENV")
	if env != "" {
		envConfig, err := box.Find(env + ".yaml")
		if err != nil {
			panic(err)
		}
		viper.SetConfigType(configType)
		err = viper.ReadConfig(bytes.NewReader(envConfig))
		if err != nil {
			panic(err)
		}
	}
	err = viper.AddRemoteProvider("consul", viper.GetString("consul.host")+":"+viper.GetString("consul.port"), "/dev/uki-goFrame")
	err = viper.ReadRemoteConfig()
	if err != nil {
		fmt.Println(err)
	}
	go func() {
		for {
			time.Sleep(time.Second * 1)
			_ = viper.WatchRemoteConfig()
		}
	}()
}