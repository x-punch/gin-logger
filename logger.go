package logger

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config represents logger configuration.
type Config struct {
	Level         string
	SkipMethods   []string
	SkipURLs      []string
	SkipURLRegexp *regexp.Regexp
}

// DefaultLogger represents default gin logger
func DefaultLogger() gin.HandlerFunc {
	return Logger(Config{Level: "DEBUG"})
}

// Logger initializes the logging middleware.
func Logger(config Config) gin.HandlerFunc {
	var level = zapcore.DebugLevel
	if len(config.Level) > 0 {
		if err := level.Set(config.Level); err != nil {
			panic(err)
		}
	}
	logger, _ := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		DisableCaller:     true,
		DisableStacktrace: true,
		Development:       false,
		Encoding:          "console",
		EncoderConfig:     zap.NewDevelopmentEncoderConfig(),
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
	}.Build()

	var skipURLs map[string]struct{}
	if length := len(config.SkipURLs); length > 0 {
		skipURLs = make(map[string]struct{}, length)
		for _, u := range config.SkipURLs {
			skipURLs[u] = struct{}{}
		}
	}
	var skipMethods map[string]struct{}
	if length := len(config.SkipMethods); length > 0 {
		skipMethods = make(map[string]struct{}, length)
		for _, m := range config.SkipMethods {
			skipMethods[m] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		start := time.Now()

		path := c.Request.URL.Path
		if len(c.Request.URL.RawQuery) > 0 {
			path = c.Request.URL.Path + "?" + c.Request.URL.RawQuery
		}
		url := c.Request.URL.EscapedPath()
		for _, p := range c.Params {
			url = strings.Replace(path, p.Value, ":"+p.Key, 1)
		}

		c.Next()

		if _, ok := skipMethods[c.Request.Method]; ok {
			return
		}
		if _, ok := skipURLs[url]; ok {
			return
		}
		if config.SkipURLRegexp != nil && config.SkipURLRegexp.MatchString(url) {
			return
		}

		msg := "Request"
		if len(c.Errors) > 0 {
			msg = strings.Join(c.Errors.Errors(), ";")
		}

		fields := []zap.Field{
			zap.Int("s", c.Writer.Status()),
			zap.String("m", c.Request.Method),
			zap.String("p", path),
			zap.String("i", c.ClientIP()),
			zap.Duration("l", time.Now().Sub(start)),
			// zap.String("ua", c.Request.UserAgent())
		}
		switch {
		case c.Writer.Status() >= http.StatusBadRequest && c.Writer.Status() < http.StatusInternalServerError:
			logger.Warn(msg, fields...)
		case c.Writer.Status() >= http.StatusInternalServerError:
			logger.Error(msg, fields...)
		default:
			logger.Info(msg, fields...)
		}
	}
}
