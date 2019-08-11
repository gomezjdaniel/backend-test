package main

import (
	"fmt"
	"time"

	"github.com/apex/log"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"upper.io/db.v3/lib/sqlbuilder"
	"upper.io/db.v3/postgresql"
)

const schema = `
CREATE TABLE IF NOT EXISTS players (
    player_id SERIAL PRIMARY KEY,
    display_name TEXT NOT NULL DEFAULT '',
    number SMALLINT NOT NULL DEFAULT 0,
    position SMALLINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS lineups (
    lineup_id SERIAL PRIMARY KEY,
    is_local BOOL NOT NULL DEFAULT FALSE,
    formation SMALLINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS lineup_players (
    lineup_id SERIAL NOT NULL REFERENCES lineups(lineup_id) ON DELETE CASCADE,
    player_id SERIAL NOT NULL REFERENCES players(player_id) ON DELETE CASCADE,
    PRIMARY KEY(lineup_id, player_id)
);`

type config struct {
	databaseURL  string
	redisURL     string
	address      string
	level        int
	disableCache bool
}

type Option func(*server)

func EnableWebLogger(s *server) {
	s.web.Use(middleware.Logger())
}

type server struct {
	web *echo.Echo
	db  sqlbuilder.Database

	config config
}

func newServer(config config, opts ...Option) (*server, error) {
	connectionURL, err := postgresql.ParseURL(config.databaseURL)
	if err != nil {
		return nil, err
	}

	sess, err := postgresql.Open(connectionURL)
	if err != nil {
		return nil, err
	}

	s := &server{
		web:    echo.New(),
		db:     sess,
		config: config,
	}

	for _, opt := range opts {
		opt(s)
	}

	log.SetLevel(log.Level(config.level))

	redisOpts, err := redis.ParseURL(config.redisURL)
	if err != nil {
		return nil, err
	}

	redisConn := redis.NewClient(redisOpts)

	_, err = redisConn.Ping().Result()
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to redis instance: %s", err)
	}

	/*s.web.GET("/test-cache", func(c echo.Context) error {
		return c.JSON(http.StatusOK, &struct {
			Timestamp int64
		}{
			Timestamp: time.Now().UTC().Unix(),
		})
	}, cache(redisConn, time.Duration(time.Second)*15))

	s.web.POST("/test-cache", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, invalidate(redisConn))*/

	s.web.POST("/players", s.createPlayer)
	s.web.GET("/players", s.listPlayers, cache(s.config.disableCache, redisConn, time.Duration(time.Second)*5))
	s.web.PUT("/players/:player_id", s.updatePlayer, playerID)
	s.web.DELETE("/players/:player_id", s.deletePlayer, playerID)

	s.web.POST("/lineups", s.createLineup)
	s.web.GET("/lineups/:lineup_id", s.getLineup, lineupID, cache(s.config.disableCache, redisConn, time.Duration(time.Second)*10))
	s.web.PUT("/lineups/:lineup_id", s.updateLineup, lineupID, invalidate(s.config.disableCache, redisConn))
	s.web.DELETE("/lineups/:lineup_id", s.deleteLineup, lineupID, invalidate(s.config.disableCache, redisConn))

	s.web.POST("/lineups/:lineup_id/players", s.addPlayerToLineup, lineupID)
	s.web.DELETE("/lineups/:lineup_id/players", s.deletePlayerFromLineup, lineupID)

	return s, nil
}

func (s *server) start() {
	s.web.Logger.Fatal(s.web.Start(s.config.address))
}

func (s *server) init() error {
	err := s.db.Tx(nil, func(tx sqlbuilder.Tx) error {
		_, err := tx.Exec(schema)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}
