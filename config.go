package main

import (
	"log"
	"time"

	"github.com/BurntSushi/toml"
)

var config Config

type dbInfo struct {
	Host     string
	User     string
	Password string
	Db       string
	Port     string
}

type general struct {
	SecretKey      string
	Host           string
	Hours          time.Duration
	GrpcAddr       string
	Proxy          string
	OpsDingWebHook string
	OpsDingKey     string
	WikiUrl        string
	WikiBaseAuth   string
	WeixiUrl       string
	WeixiBaseUrl   string
}

type Config struct {
	General general
	Mysql   dbInfo
}

func Configinit() (err error) {

	_, err = toml.DecodeFile("conf.toml", &config)

	if err != nil {
		log.Println(err.Error())
		return
	}

	return

}
