package slack

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

type RedisConfig struct {
	Addr     string
	User     string
	Password string
}

type redisStorage struct {
	memory  ThreadsStorage
	client  *redis.Client
	channel string
}

func (s *redisStorage) threadKey(threadTS string) string {
	return s.channel + ":" + threadTS + ":thread"
}

func (s *redisStorage) LookupThread(threadTS string) *IntermediatePost {
	rootPost := s.memory.LookupThread(threadTS)
	if rootPost != nil {
		return rootPost
	}
	data, err := s.client.Get(context.TODO(), s.threadKey(threadTS)).Result()
	if err != nil || len(data) == 0 {
		return nil
	}
	var result IntermediatePost
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		log.Errorf("could not unmarshal root post from redis: %v", err)
		return nil
	}
	log.Printf("Found thread root post for thread %s in redis for channel %s", threadTS, s.channel)
	s.memory.StoreThread(threadTS, &result)
	return &result
}

// nolint:gosimple
func (s *redisStorage) HasThread(threadTS string) bool {
	if s.memory.HasThread(threadTS) {
		return true
	}
	// TODO: this method should go to redis, but right now it is only used for warnings.
	return false
}

func (s *redisStorage) StoreThread(threadTS string, rootPost *IntermediatePost) {
	s.memory.StoreThread(threadTS, rootPost)
	strippedPost := *rootPost
	strippedPost.Replies = nil
	strippedPost.Attachments = nil
	postJson, err := json.Marshal(&strippedPost)
	if err != nil {
		log.Errorf("could not marshal stripped post: %v", err)
		return
	}

	if err := s.client.Set(context.TODO(), s.threadKey(threadTS), postJson, 0).Err(); err != nil {
		log.Errorf("could not store stripped post %s: %v", threadTS, err)
	}
}

func (s *redisStorage) GetChangedThreads() []*IntermediatePost {
	return s.memory.GetChangedThreads()
}

type redisFactory struct {
	client *redis.Client
}

func newRedisFactory(cfg *RedisConfig) (*redisFactory, error) {
	opts := &redis.Options{Addr: cfg.Addr, Username: cfg.User, Password: cfg.Password}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis failure: %w", err)
	}
	return &redisFactory{
		client: client,
	}, nil
}

func (s *redisFactory) newRedisStorage(channel string) ThreadsStorage {
	return &redisStorage{
		memory:  newMemoryStorage(),
		client:  s.client,
		channel: channel,
	}
}
