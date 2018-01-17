package goddd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hyperpilotio/node-agent/pkg/snap"
	"github.com/mailru/easyjson"
	dto "github.com/prometheus/client_model/go"
)

var (
	cachePath = "/tmp/snap-plugin-collector-goddd-cache.json"
)

func readCache() (*CacheType, error) {
	file, err := ioutil.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file %s msg %s", cachePath, err.Error())
	}
	var cache CacheType
	err = easyjson.Unmarshal(file, &cache)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse json from file %s msg %s", cachePath, err.Error())

	}
	if cache.CounterType == nil {
		return NewCache(), nil
	}
	return &cache, nil
}

func updateCache(cache *CacheType) error {
	buf, err := easyjson.Marshal(cache)
	if err != nil {
		return fmt.Errorf("Unable to marshal cache %s", err.Error())
	}
	err = ioutil.WriteFile(cachePath, buf, 0644)
	if err != nil {
		return fmt.Errorf("Unable to write cache %s", err.Error())
	}
	return err
}

// NewCache return an object of CacheType
func NewCache() *CacheType {
	return &CacheType{
		CounterType: make(map[string]CounterCache),
	}
}

func initCache() error {
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		cache := NewCache()
		buf, err := easyjson.Marshal(cache)
		if err != nil {
			return fmt.Errorf("Unable to parse json %s", err.Error())
		}
		err = ioutil.WriteFile(cachePath, buf, 0644)
		if err != nil {
			return fmt.Errorf("Unable to write json file %s", err.Error())
		}
	}
	return nil
}

type metricWithType struct {
	snap.Metric
	Type dto.MetricType
}

func eliminateType(mts []metricWithType) []snap.Metric {
	var metrics []snap.Metric
	for _, m := range mts {
		temp := snap.Metric{
			Namespace:   m.Namespace,
			Data:        m.Data,
			Config:      m.Config,
			Description: m.Description,
			Timestamp:   m.Timestamp,
			Version:     m.Version,
			Unit:        m.Unit,
			Tags:        m.Tags,
		}

		metrics = append(metrics, temp)
	}
	return metrics
}

func (c *GodddCollector) _cache(mts []metricWithType) ([]snap.Metric, error) {
	var err error

	c.cache, err = readCache()
	if err != nil {
		return eliminateType(mts), fmt.Errorf("Unable to resume cache %s", err)
	}

	for idx, metric := range mts {
		switch metric.Type {
		case dto.MetricType_COUNTER:
			keyForCache, exist := generateCacheKey(metric.Namespace, metric.Tags)
			if !exist {
				break
			}

			if cache, ok := c.cache.CounterType[keyForCache]; ok {
				c.cache.CounterType[keyForCache] = CounterCache{Pre: metric.Data.(float64)}
				metric.Data = metric.Data.(float64) - cache.Pre
				mts[idx] = metric
			} else {
				// add cache of this counter
				c.cache.CounterType[keyForCache] = CounterCache{Pre: metric.Data.(float64)}
				metric.Data = float64(0)
				mts[idx] = metric
			}
		}
	}

	err = updateCache(c.cache)
	if err != nil {
		return eliminateType(mts), fmt.Errorf("Failed to call updateCache %s", err.Error())
	}
	return eliminateType(mts), nil

}

func generateCacheKey(ns snap.Namespace, tags map[string]string) (string, bool) {
	if val, ok := tags["method"]; ok {
		return fmt.Sprintf("%s_method=%s", strings.Join(ns.Strings(), "/"), val), true
	}

	return "", false
}
