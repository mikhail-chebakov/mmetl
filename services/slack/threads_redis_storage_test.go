package slack

import (
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestRedisStorage(t *testing.T) {
	redis, err := miniredis.Run()
	assert.NoError(t, err)

	redisCfg := &RedisConfig{
		Addr: redis.Addr(),
	}
	factory, err := newRedisFactory(redisCfg)
	assert.NoError(t, err)

	t.Run("store, lookup post", func(t *testing.T) {
		storage := factory.newRedisStorage("channel", "")

		threadTS := "11"
		post := &IntermediatePost{
			Message: "msg",
		}
		assert.False(t, storage.HasThread(threadTS))
		assert.Nil(t, storage.LookupThread(threadTS))

		storage.StoreThread(threadTS, post)

		assert.True(t, storage.HasThread(threadTS))
		assert.NotNil(t, storage.LookupThread(threadTS))
		assert.Equal(t, "msg", storage.LookupThread(threadTS).Message)
	})

	t.Run("lookup post from another storage", func(t *testing.T) {
		storage := factory.newRedisStorage("channel", "")

		threadTS := "21"
		post := &IntermediatePost{
			Message: "msg",
		}
		storage.StoreThread(threadTS, post)
		storage.StoreThread("22", post)
		assert.Equal(t, 2, len(storage.GetChangedThreads()))

		anotherStorage := factory.newRedisStorage("channel", "")
		assert.NotNil(t, anotherStorage.LookupThread(threadTS))
		assert.Equal(t, "msg", anotherStorage.LookupThread(threadTS).Message)
		assert.Equal(t, 1, len(anotherStorage.GetChangedThreads())) // only the post that was looked up should be marked as changed
	})

	t.Run("post should retain replies", func(t *testing.T) {
		storage := factory.newRedisStorage("channel", "")

		threadTS := "31"
		post := &IntermediatePost{
			Message: "msg",
		}
		storage.StoreThread(threadTS, post)

		assert.Equal(t, 0, len(storage.LookupThread(threadTS).Replies))
		post = storage.LookupThread(threadTS)
		post.Replies = append(post.Replies, &IntermediatePost{})
		assert.Equal(t, 1, len(storage.LookupThread(threadTS).Replies))
	})

	t.Run("should strip attachments from threads", func(t *testing.T) {
		storage := factory.newRedisStorage("channel", "my_dir/")

		threadTS := "41"
		post := &IntermediatePost{
			Message: "msg",
			Attachments: []string{
				"a",
				"my_dir/some_file",
				"b",
			},
		}
		storage.StoreThread(threadTS, post)

		storage = factory.newRedisStorage("channel", "my_dir/")
		thread := storage.LookupThread(threadTS)
		assert.Equal(t, []string{"a", "b"}, thread.Attachments)
	})
}
