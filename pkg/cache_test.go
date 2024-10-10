package main

import (
	"testing"

	"github.com/allegro/bigcache/v3"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/require"
)

func TestCacheCreationUseNoCache(t *testing.T) {
	qc := pluginConfig{UseCaching: false}
	cache, err := newQueryCache(qc)
	require.NoError(t, err, "", "")
	require.Equal(t, cache, (*bigcache.BigCache)(nil))
}

func TestCacheCreationUseCache(t *testing.T) {
	qc := pluginConfig{UseCaching: true}
	cache, err := newQueryCache(qc)
	require.NoError(t, err, "", "")
	require.Equal(t, 0, cache.Len())
	err = setQueryInCache(cache, queryConfigStruct{FinalQuery: "Select 1;", CacheState: cacheState{Use: true}}, data.NewFrame(""))
	require.NoError(t, err, "", "")
	require.Equal(t, 1, cache.Len())
	frame, err := getQueryFromCache(cache, queryConfigStruct{FinalQuery: "Select 1;", CacheState: cacheState{Use: true}})
	require.NoError(t, err, "", "")
	require.Equal(t, data.NewFrame(""), frame)
	frame, err = getQueryFromCache(cache, queryConfigStruct{FinalQuery: "Select 2;", CacheState: cacheState{Use: true}})
	require.Error(t, err, "Entry not found")
	require.Equal(t, data.NewFrame(""), frame)
}

//TODO add more tests
