package config

import (
	"encoding/json"
	"flag"
	"golang.org/x/time/rate"
	"log"
	"os"
	"sync"
)

type Configuration struct {
	DatabaseHost     string `json:"host"`
	DatabasePort     string `json:"databasePort"`
	DriverName       string `json:"driverName"`
	DatabaseName     string `json:"databaseName"`
	DatabaseUser     string
	DatabasePassword string
	limiter          *rate.Limiter
}

var lock = &sync.Mutex{}
var conf *Configuration

func newConf() *Configuration {
	var newConf = &Configuration{}
	newConf.limiter = rate.NewLimiter(1, 1)
	flag.StringVar(&newConf.DatabaseUser, "user", "", "user for connect to database")
	flag.StringVar(&newConf.DatabasePassword, "pwd", "", "password")
	flag.Parse()

	b, err := os.ReadFile("./db_conf.json")
	if err != nil {
		log.Print("error when read file \"db_conf.json\". Check that this file exist and app has permission")
		return newConf
	}

	if err := json.Unmarshal(b, &newConf); err != nil {
		log.Printf("error when decode: %v", err)
		return newConf
	}

	return newConf
}

func GetInstance() *Configuration {
	if conf == nil {
		lock.Lock()
		defer lock.Unlock()
		if conf == nil {
			conf = newConf()
		}
	}
	return conf
}

func GetDriverName() string {
	return conf.DriverName
}

func GetHost() string {
	return conf.DatabaseHost
}

func GetPort() string {
	return conf.DatabasePort
}

func GetDbName() string {
	return conf.DatabaseName
}

func GetUser() string {
	return conf.DatabaseUser
}

func GetPassword() string {
	return conf.DatabasePassword
}

func GetLimiter() *rate.Limiter {
	return conf.limiter
}
