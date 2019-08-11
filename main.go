package main

import (
	"flag"

	"github.com/apex/log"
)

var (
	conf          = config{}
	defaultConfig = config{
		databaseURL:  "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
		redisURL:     "redis://:@localhost:6379/0",
		address:      ":1323",
		level:        1,
		disableCache: false,
	}
)

func main() {
	flag.StringVar(&conf.databaseURL, "database-url", defaultConfig.databaseURL, "The database URL to connect to.")
	flag.StringVar(&conf.redisURL, "redis-url", defaultConfig.redisURL, "The redis URL to connect to.")
	flag.StringVar(&conf.address, "address", defaultConfig.address, "Address the HTTP server will listen to.")
	flag.IntVar(&conf.level, "log-level", defaultConfig.level, "Log level (0-5).")
	flag.BoolVar(&conf.disableCache, "disable-cache", defaultConfig.disableCache, "Whether cache should be disabled or not.")

	flag.Parse()

	s, err := newServer(conf, EnableWebLogger)
	if err != nil {
		log.WithError(err).Fatal("Failed to create server")
	}

	err = s.init()
	if err != nil {
		log.WithError(err).Fatal("Failed to create database schema")
	}

	s.start()
}
