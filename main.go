package main

import (
	"fmt"
	"github.com/spf13/viper"
	_ "goFrame/conf"
)

func main() {
	//viper.SetEnvPrefix("ENV")
	// 读取默认配置

	//viper.SetConfigName("loc")
	//viper.AddConfigPath("conf")
	//_ = viper.ReadInConfig()
	logLevel := viper.GetString("logLevel")
	app := viper.GetString("app")
	fmt.Println(logLevel)
	fmt.Println(app)
}
