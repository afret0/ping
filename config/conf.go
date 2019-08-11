package config

import (
	"bytes"
	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"goFrame/libs"
	"io"
	"io/ioutil"
	"time"
)

type redis struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	Pwd  string `json:"pwd"`
}
type mongo struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
	Pwd  string `json:"pwd"`
}
type etc struct {
	App   string `json:"app"`
	Port  string `json:"port"`
	Ping  string `json:"ping"`
	Redis redis  `json:"redis"`
	Mongo mongo  `json:"mongo"`
}

var conf *etc

func init() {
	conf = new(etc)
	initConfig()
	updateConfig()

}

// GetConf ...
func GetConf() *etc {
	return conf
}

func updateConfig() {
	err := viper.Unmarshal(conf)
	if err != nil {
		logger := libs.GetLogger()
		logger.Info(err)
	}
}
func initConfig() {
	logger := libs.GetLogger()
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
	//err = v.ReadConfig(bytes.NewReader().)
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
	if err != nil {
		logger.Info(err)
	}
	err = viper.ReadRemoteConfig()
	if err != nil {
		//fmt.Println(err)
		logger.Info(err)
	}
	err = viper.WatchRemoteConfig()
	if err != nil {
		//fmt.Println(err)
		logger.Info(err)
	}
	updateConfig()
	go func() {
		for {
			err = viper.WatchRemoteConfig()
			if err != nil {
				logger.Info(err)
			}
			updateConfig()
			time.Sleep(time.Second * 5)
		}
	}()
}
