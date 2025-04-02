package main

import (
	"github.com/redis/go-redis/v9"
	"lucathurm.dev/tofuh/internal/options"
)

func main() {
	config := options.LoadOptions()

	db := redis.NewClient(&redis.Options{Addr: config.DbAddress, Password: config.DbPassword})
}
