package xcutil

import (
	"sync"
	"time"
)

const BucketSize = 1 << 8

var localCache *LocalCache
var once sync.Once

type LocalCache struct {
	cacheBuckets [BucketSize]*sync.Map

	expires *sync.Map // key:expireTime

	// hash function to determine which bucket to store the key(e.g. fnv32 crc32)
	hashFunc func(key string) int
}

// expire e.g. 30 * time.Minute
func GetLocalCache(expire time.Duration) *LocalCache {
	once.Do(func() {
		localCache = &LocalCache{
			expires: &sync.Map{},
		}
		for i := 0; i < BucketSize; i++ {
			localCache.cacheBuckets[i] = &sync.Map{}
		}

		go localCache.expireData(expire)
	})
	return localCache
}

func (c *LocalCache) Set(key string, val interface{}) {
	c.cacheBuckets[c.hashFunc(key)].Store(key, val)
	c.expires.Store(key, time.Now())
}

func (c *LocalCache) Get(key string) (interface{}, bool) {
	return c.cacheBuckets[c.hashFunc(key)].Load(key)
}

func (c *LocalCache) expireData(expire time.Duration) {
	for {
		c.expires.Range(func(key, expireTime interface{}) bool {
			if time.Since(expireTime.(time.Time)) > expire {
				c.delete(key.(string))
			}
			return true
		})
		time.Sleep(2 * time.Minute)
	}
}

func (c *LocalCache) delete(key string) {
	c.cacheBuckets[c.hashFunc(key)].Delete(key)
}
