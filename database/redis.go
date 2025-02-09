package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background() // Define global context

const SessionExpiration = time.Hour * 24

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func SetSession(userID string, token string) error {
	return RedisClient.Set(Ctx, "session:"+userID, token, SessionExpiration).Err()
}

func GetSession(userID string) (string, error) {
	return RedisClient.Get(Ctx, "session:"+userID).Result()
}

func DeleteSession(userID string) error {
	return RedisClient.Del(Ctx, "session:"+userID).Err()
}

func TrackActiveUser(userID string) error {
	return RedisClient.SAdd(Ctx, "active_users", userID).Err()
}

func RemoveActiveUser(userID string) error {
	return RedisClient.SRem(Ctx, "active_users", userID).Err()
}

func GetActiveUsers() ([]string, error) {
	return RedisClient.SMembers(Ctx, "active_users").Result()
}
