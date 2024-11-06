package utils

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// Limiters map untuk menyimpan rate limiter per IP
var limiters = make(map[string]*rate.Limiter)

// createLimiter membuat limiter baru jika belum ada untuk IP tertentu
func createLimiter(rps float64, burst int) *rate.Limiter {
	return rate.NewLimiter(rate.Limit(rps), burst)
}

// getLimiter mengembalikan rate limiter untuk IP tertentu
func getLimiter(ip string) *rate.Limiter {
	if limiter, exists := limiters[ip]; exists {
		return limiter
	}

	limiter := createLimiter(3, 5) // Limit 3 request per second, burst 5
	limiters[ip] = limiter

	// Bersihkan limiter untuk IP yang tidak aktif setelah periode tertentu
	go func() {
		time.Sleep(5 * time.Minute)
		delete(limiters, ip)
	}()
	return limiter
}

// rateLimitMiddleware adalah middleware untuk membatasi rate request per IP di Echo
func RateLimitMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ip := c.RealIP()

		limiter := getLimiter(ip)
		if !limiter.Allow() {
			return c.JSON(http.StatusTooManyRequests,
				map[string]interface{}{"code": "LOS-KMB-429", "data": nil, "errors": "TooManyRequests", "message": "Terlalu banyak request, silahkan coba lagi", "server_time": GenerateTimeNow, "x-request-id": echo.HeaderXRequestID})
		}

		return next(c)
	}
}
