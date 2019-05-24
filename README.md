# gin-logger
Gin middleware used to logger url request, which support advanced filter.

# Usage
```
app.Use(logger.DefaultLogger())
```
```
app.Use(logger.Logger(logger.Config{Level: "WARN", SkipMethods: []string{"OPTIONS"}, SkipURLs: []string{"/test"}, SkipURLRegexp: regexp.MustCompile("/swagger/*")}))
```
