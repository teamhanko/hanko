package rate_limiter

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/gomodule/redigo/redis"
	"github.com/labstack/echo/v4"
	"github.com/sethvargo/go-limiter"
	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
	"github.com/sethvargo/go-redisstore"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

func NewRateLimiter(cfg config.RateLimiter, limits config.RateLimits) limiter.Store {
	if cfg.Store == config.RATE_LIMITER_STORE_REDIS {
		store, err := redisstore.New(&redisstore.Config{
			Tokens:   limits.Tokens,
			Interval: limits.Interval,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", cfg.Redis.Address,
					redis.DialPassword(cfg.Redis.Password))
			},
		})
		if err != nil {
			log.Fatal(err)
		}
		return store
	}
	// else return in_memory
	store, err := memorystore.New(&memorystore.Config{
		Tokens:   limits.Tokens,
		Interval: limits.Interval,
	})
	if err != nil {
		log.Fatal(err)
	}
	return store
}

func Limit(store limiter.Store, userId uuid.UUID, c echo.Context) error {
	key := c.Path() + "/" + userId.String() + "/" + c.RealIP()
	// Take from the store.
	limit, remaining, reset, ok, err := store.Take(context.Background(), key)

	if err != nil {
		return err
	}

	resetTime := int(math.Floor(time.Unix(0, int64(reset)).UTC().Sub(time.Now().UTC()).Seconds()))
	log.Println(resetTime)

	// Set headers (we do this regardless of whether the request is permitted).
	c.Response().Header().Set(httplimit.HeaderRateLimitLimit, strconv.FormatUint(limit, 10))
	c.Response().Header().Set(httplimit.HeaderRateLimitRemaining, strconv.FormatUint(remaining, 10))
	c.Response().Header().Set(httplimit.HeaderRateLimitReset, strconv.Itoa(resetTime))

	// Fail if there were no tokens remaining.
	if !ok {
		c.Response().Header().Set(httplimit.HeaderRetryAfter, strconv.Itoa(resetTime))
		return dto.NewHTTPError(http.StatusTooManyRequests)
	}
	return nil
}
