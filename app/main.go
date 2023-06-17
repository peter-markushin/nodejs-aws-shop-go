package main

import (
	"github.com/peterm-itr/nodejs-aws-shop-go/config"
	"github.com/peterm-itr/nodejs-aws-shop-go/db"
	"github.com/peterm-itr/nodejs-aws-shop-go/repositories"
	"github.com/peterm-itr/nodejs-aws-shop-go/server"
	"log"
)

func main() {
	configuration, err := config.GetConfig()

	if err != nil {
		log.Println(err.Error())
	}

	db.Init(configuration)
	log.Println("DB Init complete")

	repositories.Init()
	log.Println("Repo Init complete")

	server.Init(configuration)
	log.Println("Server Init complete")
}
