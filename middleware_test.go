package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	r := require.New(t)

	redisURL := strings.Replace(defaultConfig.redisURL, "@redis", "@localhost", 1)

	opts, err := redis.ParseURL(redisURL)
	r.Nil(err)

	conn := redis.NewClient(opts)

	_, err = conn.Ping().Result()
	r.Nil(err)

	ttl := time.Duration(time.Second) * 5

	counter := 0

	web := echo.New()
	web.GET("/test", func(c echo.Context) error {
		defer func() {
			counter = counter + 1
		}()

		return c.String(http.StatusOK, fmt.Sprintf("%d", counter))
	}, cache(false, conn, ttl))
	web.POST("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, invalidate(false, conn))

	{
		rec := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/test", nil)
		r.Nil(err)

		web.ServeHTTP(rec, req)

		data, err := ioutil.ReadAll(rec.Result().Body)
		r.Nil(err)
		r.Equal("0", string(data))
	}

	{
		rec := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/test", nil)
		r.Nil(err)

		web.ServeHTTP(rec, req)

		data, err := ioutil.ReadAll(rec.Result().Body)
		r.Nil(err)
		r.Equal("0", string(data))
	}

	{
		rec := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/test", nil)
		r.Nil(err)

		web.ServeHTTP(rec, req)

		data, err := ioutil.ReadAll(rec.Result().Body)
		r.Nil(err)
		r.Equal("0", string(data))
	}

	time.Sleep(ttl)

	{
		rec := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/test", nil)
		r.Nil(err)

		web.ServeHTTP(rec, req)

		data, err := ioutil.ReadAll(rec.Result().Body)
		r.Nil(err)
		r.Equal("1", string(data))
	}

	{
		rec := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/test", nil)
		r.Nil(err)

		web.ServeHTTP(rec, req)

		data, err := ioutil.ReadAll(rec.Result().Body)
		r.Nil(err)
		r.Equal("1", string(data))
	}

	// Invalidate cache.

	{
		rec := httptest.NewRecorder()

		req, err := http.NewRequest("POST", "/test", nil)
		r.Nil(err)

		web.ServeHTTP(rec, req)
	}

	{
		rec := httptest.NewRecorder()

		req, err := http.NewRequest("GET", "/test", nil)
		r.Nil(err)

		web.ServeHTTP(rec, req)

		data, err := ioutil.ReadAll(rec.Result().Body)
		r.Nil(err)
		r.Equal("2", string(data))
	}
}
