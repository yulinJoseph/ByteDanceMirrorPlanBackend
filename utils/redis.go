package utils

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v9"
	"github.com/spf13/viper"
	"path"
	"runtime"
	"time"
)

var Redis *redis.Client

func InitRedis() {
	_, fileName, _, ok := runtime.Caller(0)
	if !ok {
		fmt.Println("No caller information")
		return
	}
	fileName = path.Join(path.Dir(fileName), "../config/redis.json")
	viper.SetConfigFile(fileName)

	Redis = redis.NewClient(&redis.Options{
		Addr:         viper.GetString("addr"),
		Password:     viper.GetString("password"),
		DB:           viper.GetInt("DB"),
		PoolSize:     viper.GetInt("poolSize"),
		MinIdleConns: viper.GetInt("minIdleConns"),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if pong, err := Redis.Ping(ctx).Result(); err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("ping redis: %s\n", pong)
	}
}
