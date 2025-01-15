package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/zhengxiaod/gochat/pkg/db"
	"github.com/zhengxiaod/gochat/router"
	"io"
	"log"
	"os"
)

func main() {
	// 读取config/app.yaml文件配置
	initConfig()

	// 初始化mysql数据库
	db.InitMySQL()

	go func() {
		e := router.HttpRouter()
		e.Run(":8080")
	}()

	r := router.WSRouter()
	r.Run(":8081")

	// 将gin框架的日志输出到log/gin.log文件中
	//initLogFile()
}

func initConfig() {
	viper.SetConfigName("config/app")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func initLogFile() {
	// Disable Console Color, you don't need console color when writing the logs to file.
	gin.DisableConsoleColor()

	// Logging to a file.
	logFile := viper.GetString("app.logFile")
	f, err := os.Create(logFile)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	gin.DefaultWriter = io.MultiWriter(f)
}
