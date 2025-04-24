package config

import (
	"os"
	"strconv"
	"time"
)

const (
	dbAddressKey       = "DB_ADDRESS"
	dbPasswordKey      = "DB_PASSWORD"
	dbStateKeyKey      = "DB_STATE_KEY"
	dbTimeoutKey       = "DB_TIMEOUT"
	msgAddressKey      = "MSG_ADDRESS"
	msgClientIDKey     = "MSG_CLIENT_ID"
	msgConnTimeoutKey  = "MSG_CONN_TIMEOUT"
	msgQOSKey          = "MSG_QOS"
	msgRetainStateKey  = "MSG_RETAIN_STATE"
	msgStateTopicKey   = "MSG_STATE_TOPIC"
	msgSetBitTopicKey  = "MSG_SET_BIT_TOPIC"
	httpAddressKey     = "HTTP_ADDRESS"
	httpClientsFileKey = "HTTP_CLIENTS_FILE"
)

const (
	dbAddressFallback      = "database:6379"
	dbStateKeyFallback     = "state"
	dbTimeoutFallback      = 1000 * time.Millisecond
	msgAddressFallback     = "tcp://broker:1883"
	msgClientIDFallback    = "tofuh-app"
	msgConnTimeoutFallback = 3000 * time.Millisecond
	msgQOSFallback         = 1
	msgRetainStateFallback = true
	msgStateTopicFallback  = "tofuh/state"
	msgSetBitTopicFallback = "tofuh/setbit"
	httpAddressFallback    = ":80"
	httpClientFileFallback = "/secrets/httpclients.json"
)

type Config struct {
	DbAddress         string
	DbPassword        string
	DbStateKey        string
	DbTimeout         time.Duration
	MsgAddress        string
	MsgClientID       string
	MsgConnTimeout    time.Duration
	MsgQOS            byte
	MsgRetainState    bool
	MsgStateTopicKey  string
	MsgSetBitTopicKey string
	HTTPAddress       string
	HTTPClientsFile   string
}

func LoadConfig() Config {
	return Config{
		DbAddress:         getEnvStr(dbAddressKey, dbAddressFallback),
		DbPassword:        getEnvStr(dbPasswordKey, ""),
		DbStateKey:        getEnvStr(dbStateKeyKey, dbStateKeyFallback),
		DbTimeout:         getEnvDuration(dbTimeoutKey, dbTimeoutFallback),
		MsgAddress:        getEnvStr(msgAddressKey, msgAddressFallback),
		MsgClientID:       getEnvStr(msgClientIDKey, msgClientIDFallback),
		MsgConnTimeout:    getEnvDuration(msgConnTimeoutKey, msgConnTimeoutFallback),
		MsgQOS:            getEnvByte(msgQOSKey, msgQOSFallback),
		MsgRetainState:    getEnvBool(msgRetainStateKey, msgRetainStateFallback),
		MsgStateTopicKey:  getEnvStr(msgStateTopicKey, msgStateTopicFallback),
		MsgSetBitTopicKey: getEnvStr(msgSetBitTopicKey, msgSetBitTopicFallback),
		HTTPAddress:       getEnvStr(httpAddressKey, httpAddressFallback),
		HTTPClientsFile:   getEnvStr(httpClientsFileKey, httpClientFileFallback),
	}
}

func getEnvStr(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	str := getEnvStr(key, "")
	i, err := strconv.Atoi(str)
	if err != nil {
		return fallback
	}
	return time.Millisecond * time.Duration(i)
}

func getEnvByte(key string, fallback byte) byte {
	str := getEnvStr(key, "")
	b, err := strconv.ParseUint(str, 10, 8)
	if err != nil {
		return fallback
	}
	return byte(b)
}

func getEnvBool(key string, fallback bool) bool {
	str := getEnvStr(key, "")
	b, err := strconv.ParseBool(str)
	if err != nil {
		return fallback
	}
	return b
}
