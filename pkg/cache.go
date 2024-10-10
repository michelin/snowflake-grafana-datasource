package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type cacheState struct {
	Use   bool
	Until time.Time
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func newQueryCache(config pluginConfig) (*bigcache.BigCache, error) {
	if !config.UseCaching {
		return nil, nil
	}
	IntCacheSize := 0
	IntCacheRetention := 0
	if config.CacheSize == "" {
		config.CacheSize = "2048"
	}
	if config.CacheRetention == "" {
		config.CacheRetention = "60"
	}
	if CacheSize, err := strconv.Atoi(config.CacheSize); err == nil {
		IntCacheSize = int(CacheSize)
	} else {
		return nil, err
	}
	if CacheRetention, err := strconv.Atoi(config.CacheRetention); err == nil {
		IntCacheRetention = int(CacheRetention)
	} else {
		return nil, err
	}
	cache_config := bigcache.Config{
		// number of shards (must be a power of 2)
		Shards: 1024,

		// time after which entry can be evicted
		LifeWindow: time.Duration(IntCacheRetention) * time.Minute,

		// Interval between removing expired entries (clean up).
		// If set to <= 0 then no action is performed.
		// Setting to < 1 second is counterproductive — bigcache has a one second resolution.
		CleanWindow: 5 * time.Minute,

		// rps * lifeWindow, used only in initial memory allocation
		MaxEntriesInWindow: 1000 * 10 * 60,

		// max entry size in bytes, used only in initial memory allocation
		MaxEntrySize: 500,

		// prints information about additional memory allocation
		Verbose: true,

		// cache will not allocate more memory than this limit, value in MB
		// if value is reached then the oldest entries can be overridden for the new ones
		// 0 value means no size limit
		HardMaxCacheSize: IntCacheSize,

		// callback fired when the oldest entry is removed because of its expiration time or no space left
		// for the new entry, or because delete was called. A bitmask representing the reason will be returned.
		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
		OnRemove: nil,

		// OnRemoveWithReason is a callback fired when the oldest entry is removed because of its expiration time or no space left
		// for the new entry, or because delete was called. A constant representing the reason will be passed through.
		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
		// Ignored if OnRemove is specified.
		OnRemoveWithReason: nil,
	}
	cache, err := bigcache.New(context.Background(), cache_config)
	return cache, err
}

func getQueryFromCache(cache *bigcache.BigCache, queryConfig queryConfigStruct) (*data.Frame, error) {
	frame := data.NewFrame("")
	if cache == nil || !queryConfig.CacheState.Use {
		return frame, errors.New("noCache")
	}
	log.DefaultLogger.Info("Cache", queryConfig.CacheState.Until.Format(time.RFC3339))
	cache_res, err := cache.Get(GetMD5Hash(queryConfig.CacheState.Until.Format(time.RFC3339) + queryConfig.FinalQuery))
	if err != nil {
		return frame, err
	}
	log.DefaultLogger.Info("Snowflake cache hit")
	frame.UnmarshalJSON(cache_res)
	return frame, err
}

func setQueryInCache(cache *bigcache.BigCache, queryConfig queryConfigStruct, frame *data.Frame) error {
	if cache == nil || !queryConfig.CacheState.Use {
		return errors.New("noCache")
	}
	json, err := json.Marshal(frame)
	if err == nil {
		cache.Set(GetMD5Hash(queryConfig.CacheState.Until.Format(time.RFC3339)+queryConfig.FinalQuery), json)
	}
	return err
}
