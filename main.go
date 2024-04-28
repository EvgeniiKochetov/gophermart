package main

import (
	"gophermart/src"
	"gophermart/src/config"
	"gophermart/src/databases"
)

func main() {

	config.GetInstance()
	databases.GetInstance()
	src.Run()
}
