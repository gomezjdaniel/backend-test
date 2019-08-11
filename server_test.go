package main

import "fmt"

const (
	testDatabase    = "test"
	testDatabaseURL = "postgres://postgres:postgres@localhost:5432/%s?sslmode=disable"
)

func testServer() *server {
	config := defaultConfig
	config.databaseURL = fmt.Sprintf(testDatabaseURL, "template1")

	s, err := newServer(config)
	if err != nil {
		panic(err)
	}

	_, err = s.db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDatabase))
	if err != nil {
		panic(err)
	}

	_, err = s.db.Exec(fmt.Sprintf("CREATE DATABASE %s", testDatabase))
	if err != nil {
		panic(err)
	}

	err = s.db.Close()
	if err != nil {
		panic(err)
	}

	config.databaseURL = fmt.Sprintf(testDatabaseURL, testDatabase)
	config.disableCache = true

	s, err = newServer(config)
	if err != nil {
		panic(err)
	}

	err = s.init()
	if err != nil {
		panic(err)
	}

	return s
}
