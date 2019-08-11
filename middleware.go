package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/go-redis/redis"
	"github.com/labstack/echo/v4"
)

type redisEntry struct {
	Code int
	Body []byte
}

func cache(disableCache bool, conn *redis.Client, ttl time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if disableCache {
				return next(c)
			}

			u := c.Request().URL.String()

			content, err := conn.Get(u).Result()
			if content != "" {
				log.WithField("url_path", u).Debug("Cache hit")

				var entry redisEntry
				err := json.Unmarshal([]byte(content), &entry)
				if err != nil {
					return err
				}

				return c.String(entry.Code, string(entry.Body))
			}

			resBody := new(bytes.Buffer)
			c.Response().Writer = &bodyDumpResponseWriter{
				Writer:         io.MultiWriter(c.Response().Writer, resBody),
				ResponseWriter: c.Response().Writer,
			}

			if err := next(c); err != nil {
				return err
			}

			b, err := json.Marshal(&redisEntry{
				Code: c.Response().Status,
				Body: resBody.Bytes(),
			})
			if err != nil {
				return err
			}

			conn.Set(u, b, ttl)

			return nil
		}
	}
}

func invalidate(disableCache bool, conn *redis.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if disableCache {
				return next(c)
			}

			err := next(c)
			if err != nil {
				return err
			}

			conn.Del(c.Request().URL.String())

			return nil
		}
	}
}

type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *bodyDumpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}
