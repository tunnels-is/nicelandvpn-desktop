package main

import (
	"log"
	"os"
)

var FLAG_COMMAND string
var FLAG_USER string
var FLAG_PASSWORD string

func main() {

	log.Println(os.Args)
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	switch os.Args[1] {
	case "connect":
		Connect()
	case "getApiKey":
		GetAPIKEy()
	case "createConfig":
		CreateDummyConfig()
	default:
		os.Exit(1)
	}

}

func GetAPIKEy() {

}

func CreateDummyConfig() {

}
func Connect() {

}
