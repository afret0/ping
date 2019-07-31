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

	env := os.Getenv("ENV")
	if env != "" {
		envConfig, err := box.Find(env + ".yaml")
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
}
