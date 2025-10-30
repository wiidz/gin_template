package middleware

import (
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"gin_template/internal/common/response"
)

// RateLimit applies a global token-bucket limiter per process.
// rps: tokens per second; burst: maximum burst size.
func RateLimit(rps float64, burst int) gin.HandlerFunc {
	lim := rate.NewLimiter(rate.Limit(rps), burst)
	return func(c *gin.Context) {
		if !lim.Allow() {
			response.Error(c, 429, "too many requests")
			return
		}
		c.Next()
	}
}

// RateLimitIP applies per-IP token-bucket limiting.
func RateLimitIP(rps float64, burst int) gin.HandlerFunc {
	var limiters sync.Map // ip -> *rate.Limiter
	newLimiter := func() *rate.Limiter { return rate.NewLimiter(rate.Limit(rps), burst) }
	return func(c *gin.Context) {
		ip := c.ClientIP()
		v, ok := limiters.Load(ip)
		if !ok {
			lim := newLimiter()
			v, _ = limiters.LoadOrStore(ip, lim)
		}
		lim := v.(*rate.Limiter)
		if !lim.Allow() {
			response.Error(c, 429, "too many requests")
			return
		}
		c.Next()
	}
}

// In-memory IP denylist
var denyIPs sync.Map // ip -> struct{}

func DenyIP(ip string)  { denyIPs.Store(ip, struct{}{}) }
func AllowIP(ip string) { denyIPs.Delete(ip) }
func IsDenied(ip string) bool {
	_, ok := denyIPs.Load(ip)
	return ok
}

// IPDenylist blocks requests from denied IPs with 403.
func IPDenylist() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsDenied(c.ClientIP()) {
			response.Error(c, 403, "forbidden")
			return
		}
		c.Next()
	}
}
