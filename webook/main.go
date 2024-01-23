package main

import "github.com/spf13/viper"

func main() {
	initViper()
	server := initApp()
	err := server.Run(":8080")
	if err != nil {
		return
	}
}

func initViper() {
	viper.SetConfigFile("config/dev.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
