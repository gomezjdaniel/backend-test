package main

import (
	"fmt"
	"strings"

	"github.com/apex/log"
)

const (
	testDatabase    = "test"
	testDatabaseURL = "postgres://postgres:postgres@localhost:5432/%s?sslmode=disable"
)

func testServer() *server {
	config := defaultConfig
	config.databaseURL = fmt.Sprintf(testDatabaseURL, "template1")
	config.redisURL = strings.Replace(defaultConfig.redisURL, "@redis", "@localhost", 1)

	s, err := newServer(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to create server")
	}

	_, err = s.db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDatabase))
	if err != nil {
		log.WithError(err).Fatal("Failed to drop test database")
	}

	_, err = s.db.Exec(fmt.Sprintf("CREATE DATABASE %s", testDatabase))
	if err != nil {
		log.WithError(err).Fatal("Failed to create test database")
	}

	err = s.db.Close()
	if err != nil {
		log.WithError(err).Fatal("Failed to close `template1` database connection")
	}

	config.databaseURL = fmt.Sprintf(testDatabaseURL, testDatabase)
	config.disableCache = true

	s, err = newServer(config)
	if err != nil {
		log.WithError(err).Fatal("Failed to create server")
	}

	err = s.init()
	if err != nil {
		log.WithError(err).Fatal("Failed to create database schema")
	}

	return s
}
