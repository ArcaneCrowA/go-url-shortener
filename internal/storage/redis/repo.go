package redisdb

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const codePrefix = "url:"

type Repo struct {
	client *redis.Client
}

func New(client *redis.Client) *Repo {
	return &Repo{
		client: client,
	}
}

func (r *Repo) InsertRecord(code, site string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.client.Set(ctx, codePrefix+code, site, time.Hour).Err()
}

func (r *Repo) GetRecord(code string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.client.Get(ctx, codePrefix+code).Result()
}
