package main

import (
	"fmt"
	"goFrame/conf"
	"time"
)

func main() {
	//viper.SetEnvPrefix("ENV")
	// 读取默认配置

	//viper.SetConfigName("loc")
	//viper.AddConfigPath("conf")
	//_ = viper.ReadInConfig()
	//logLevel := viper.GetString("logLevel")
	//app := viper.GetString("app")
	//ping := viper.GetString("ping")
	//fmt.Println(logLevel)
	for {
		//fmt.Println(viper.GetString("ping"))
		fmt.Println(conf.Etc)
		time.Sleep(time.Second * 1)
	}
}
