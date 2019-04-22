package logger

import (
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config represents logger configuration.
type Config struct {
	UTC            bool
	Level          string
	SkipPaths      []string
	SkipPathRegexp *regexp.Regexp
}

// DefaultLogger represents default gin logger
func DefaultLogger() gin.HandlerFunc {
	return Logger(Config{UTC: true, Level: "DEBUG"})
}

// Logger initializes the logging middleware.
func Logger(config Config) gin.HandlerFunc {
	var skip map[string]struct{}
	if length := len(config.SkipPaths); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range config.SkipPaths {
			skip[path] = struct{}{}
		}
	}

	level := zapcore.DebugLevel
	if len(config.Level) > 0 {
		level.Set(config.Level)
	}
	logger, _ := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		DisableCaller:    true,
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		c.Next()
		track := true

		if _, ok := skip[path]; ok {
			track = false
		}
		if track && config.SkipPathRegexp != nil && config.SkipPathRegexp.MatchString(path) {
			track = false
		}

		if !track {
			return
		}
		end := time.Now()
		latency := end.Sub(start)
		if config.UTC {
			end = end.UTC()
		}

		msg := "Request"
		if len(c.Errors) > 0 {
			msg = c.Errors.String()
		}

		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("ip", c.ClientIP()),
			zap.Duration("latency", latency),
			zap.String("user-agent", c.Request.UserAgent())}

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
