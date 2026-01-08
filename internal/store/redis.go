package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var Ctx = context.Background()

func InitRedis(addr string) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := RDB.Ping(Ctx).Result()
	if err != nil {
		return fmt.Errorf("redis bağlantı hatası: %v", err)
	}
	return nil
}

// CheckQuota decrements the quota for an API key. Returns true if allowed.
func CheckQuota(apiKey string) (bool, error) {
	// Key format: "quota:<api_key>"
	key := fmt.Sprintf("quota:%s", apiKey)

	val, err := RDB.Decr(Ctx, key).Result()
	if err != nil {
		// If key doesn't exist, maybe set it first?
		// For now, assume keys are initialized by Admin or have default logic.
		// If we want auto-create:
		if err == redis.Nil {
			// Key missing, logic depends on requirements. Let's assume 0 quota if missing.
			return false, nil
		}
		return false, err
	}

	if val < 0 {
		return false, nil
	}
	return true, nil
}

// SetQuota sets the initial quota for a key
func SetQuota(apiKey string, limit int) error {
	key := fmt.Sprintf("quota:%s", apiKey)
	return RDB.Set(Ctx, key, limit, 24*time.Hour*30).Err() // 30 days expiry
}
