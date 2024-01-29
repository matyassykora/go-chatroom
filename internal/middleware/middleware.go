package middleware

import (
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/monitor"
)

var (
	Compress = compress.New(
		compress.Config{
			Level: compress.LevelBestSpeed,
		})

	BasicAuth = basicauth.New(basicauth.Config{
		Users: map[string]string{
			"admin": "admin",
		},
		Next: filterURL,
	})

	Helmet = helmet.New()

	Limiter = limiter.New(limiter.Config{
		Max:               20,
		Expiration:        5 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	})

	Monitor = monitor.New(monitor.Config{Title: "Performance Metrics"})
)

func filterURL(c *fiber.Ctx) bool {
	originalURL := strings.ToLower(c.OriginalURL())

	protectedURLs := []*regexp.Regexp{
		regexp.MustCompile("^/metrics$"),
	}

	for _, url := range protectedURLs {
		if url.MatchString(originalURL) {
			return false
		}
	}
	return true
}
