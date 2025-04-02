package options

import "os"

const (
	dbAddressKey  = "DB_ADDRESS"
	dbPasswordKey = "DB_PASSWORD"
	msgAddressKey = "MSG_ADDRESS"
)

const (
	dbAddressFallback  = "redis:6379"
	msgAddressFallback = "tcp://mosquitto:1883"
)

type Options struct {
	DbAddress  string
	DbPassword string
	MsgAddress string
}

func LoadOptions() Options {
	return Options{
		DbAddress:  getEnv(dbAddressKey, dbAddressFallback),
		DbPassword: getEnv(dbPasswordKey, ""),
		MsgAddress: getEnv(msgAddressKey, msgAddressFallback),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return fallback
}
