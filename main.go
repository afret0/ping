package main

import (
	"fmt"
	"github.com/spf13/viper"
	_ "goFrame/conf"
	"time"
)

func main() {
	fmt.Println(viper.GetString("app"))
	for {
		fmt.Println(viper.GetString("ping"))
		time.Sleep(time.Second * 1)
	}
}
