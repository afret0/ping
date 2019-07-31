package conf

import (
	"bytes"
	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/viper"
	"os"
)

func init() {
	box := packr.New("confBox", ".")
	configType := "yaml"
	defaultConfig, err := box.Find("app.yaml")
	if err != nil {
		panic(err)
		return
	}
	v := viper.New()
	v.SetConfigType(configType)
	err = v.ReadConfig(bytes.NewReader(defaultConfig))
	if err != nil {
		panic(err)
		return
	}
	configs := v.AllSettings()
	for k, v := range configs {
		viper.SetDefault(k, v)
	}

	//configType := "yaml"
	//defaultPath := "conf"
	//baseConfig := viper.New()
	//baseConfig.SetConfigName("app")
	//baseConfig.AddConfigPath(defaultPath)
	//baseConfig.SetConfigType(configType)
	//err := baseConfig.ReadInConfig()
	//if err != nil {
	//	//panic(fmt.Errorf("Fatal error config file: %s \n", err))
	//	panic(err)
	//	return
	//}
	//config := baseConfig.AllSettings()
	//for k, v := range config {
	//	viper.SetDefault(k, v)
	//}
	//
	env := os.Getenv("ENV")
	if env != "" {
		envConfig, err := box.Find(env+".yaml")
		if err != nil {
			panic(err)
			return
		}
		viper.SetConfigType(configType)
		err = viper.ReadConfig(bytes.NewReader(envConfig))
		if err != nil {
			panic(err)
		}
	}
	//fmt.Println(env)
	//if env != "" {
	//	viper.SetConfigName("loc")
	//	viper.AddConfigPath("conf")
	//	//viper.SetConfigType(configType)
	//	err = viper.ReadInConfig()
	//	if err != nil {
	//		panic(err)
	//		return
	//	}
	//} else {
	//	fmt.Println("env is empty")
	//	return
	//}
	//err = viper.ReadInConfig()
	//if err != nil {
	//	panic(err)
	//	return
	//}
}
