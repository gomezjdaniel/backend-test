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
		return c.String(http.StatusOK, fmt.Sprintf("%d", counter))
	}, cache(false, conn, ttl))
	web.POST("/test", func(c echo.Context) error {
		counter = counter + 1
		return c.NoContent(http.StatusOK)
	}, invalidate(false, conn))

	for _, tc := range []struct {
		Name         string
		Method       string
		ExpectedBody string
		BeforeTest   func()
	}{
		{
			Name:         "GET request",
			Method:       "GET",
			ExpectedBody: "0",
		},
		{
			Name:         "GET request with cached response",
			Method:       "GET",
			ExpectedBody: "0",
			BeforeTest: func() {
				counter += 1
			},
		},
		{
			Name:         "GET request after cache entry expired",
			Method:       "GET",
			ExpectedBody: "1",
			BeforeTest: func() {
				time.Sleep(ttl)
			},
		},
		{
			Name:   "POST request to invalidate cache",
			Method: "POST",
		},
		{
			Name:         "GET request with cached response",
			Method:       "GET",
			ExpectedBody: "2",
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			r := require.New(t)

			if tc.BeforeTest != nil {
				tc.BeforeTest()
			}

			req := httptest.NewRequest(tc.Method, "/test", nil)
			rec := httptest.NewRecorder()

			web.ServeHTTP(rec, req)

			resp := rec.Result()

			r.Equal(http.StatusOK, resp.StatusCode)

			data, err := ioutil.ReadAll(resp.Body)
			r.Nil(err)
			r.Equal(tc.ExpectedBody, strings.TrimRight(string(data), "\n"))
		})
	}
}
