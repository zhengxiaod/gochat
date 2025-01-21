package db

import (
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"log"
)

var (
	RDB *redis.Client
)

type RedisClient struct {
	Redis RedisConfig `mapstructure:"redis" yaml:"redis"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr" yaml:"addr"`         // 服务器地址
	Password string `mapstructure:"password" yaml:"password"` // 数据库密码
}

func InitRedis() {
	var err error
	// 解析配置文件
	var config RedisClient
	if err = viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to decode into struct: %v", err)
	}

	RDB = redis.NewClient(&redis.Options{
		Addr:         config.Redis.Addr,
		DB:           0,
		Password:     config.Redis.Password,
		PoolSize:     30,
		MinIdleConns: 30,
	})
	_, err = RDB.Ping().Result()
	if err != nil {
		log.Fatalf("redis connect failed, err:%v", err)
	}

}
