package db

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

var DB *gorm.DB

type MySQLClient struct {
	MySQL MySQLConfig `mapstructure:"mysql" yaml:"mysql"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host" yaml:"host"`         // 服务器地址
	Port     int    `mapstructure:"port" yaml:"port"`         // 端口
	DBName   string `mapstructure:"dbname" yaml:"dbname"`     // 数据库名
	User     string `mapstructure:"user" yaml:"user"`         // 数据库用户名
	Password string `mapstructure:"password" yaml:"password"` // 数据库密码
}

func InitMySQL() {
	var err error
	// 解析配置文件
	var config MySQLClient
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	// 构建MySQL数据源名称（DSN）
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.MySQL.User,
		config.MySQL.Password,
		config.MySQL.Host,
		config.MySQL.Port,
		config.MySQL.DBName,
	)

	//连接到数据库
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接MySQL失败: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("获取数据库连接失败: %v", err)
	}
	// SetMaxIdleConns 用于设置连接池中空闲连接的最大数量
	sqlDB.SetMaxIdleConns(20)

	// SetMaxOpenConns 设置打开数据库连接的最大数量
	sqlDB.SetMaxOpenConns(30)

	// SetConnMaxLifetime 设置了连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(time.Hour)

	fmt.Println("MySQL连接成功")
}
