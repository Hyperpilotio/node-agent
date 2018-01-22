package nodeanalyzer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/snap"
	log "github.com/sirupsen/logrus"
)

type NodeAnalyzer struct {
	DerivedMetrics      *DerivedMetrics
	NormalizerDataCache map[string]float64
	mutex               sync.Mutex
}

func init() {
	log.SetLevel(common.GetLevel(os.Getenv("SNAP_LOG_LEVEL")))
}

// NewProcessor generate processor
func NewAnalyzer() *NodeAnalyzer {
	return &NodeAnalyzer{
		NormalizerDataCache: make(map[string]float64),
	}
}

func (p *NodeAnalyzer) cacheNormalizerData(mts []snap.Metric) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, mt := range mts {
		metricNm := strings.Join(mt.Namespace.Strings(), "/")
		nodename := mt.Tags["nodename"]
		cacheKey := nodename + "/" + metricNm
		for normalizerNm, _ := range p.NormalizerDataCache {
			if strings.HasPrefix("intel/docker", normalizerNm) && strings.HasPrefix("intel/docker", metricNm) {
				dockerId := mt.Namespace.Strings()[2]
				normalizerNm = strings.Replace(normalizerNm, "*", dockerId, 1)
			}

			if metricNm == normalizerNm {
				p.NormalizerDataCache[cacheKey] = convertFloat64(mt.Data)
			}
		}
	}
}

func (p *NodeAnalyzer) ProcessMetrics(mts []snap.Metric) ([]snap.Metric, error) {
	p.cacheNormalizerData(mts)
	log.Infof("Cache normalizer data: %+v", p.NormalizerDataCache)

	metrics := []snap.Metric{}
	for _, mt := range mts {
		currentTime := mt.Timestamp.UnixNano()
		metricData := MetricData{
			MetricName:          strings.Join(mt.Namespace.Strings(), "/"),
			NodeName:            mt.Tags["nodename"],
			Value:               convertFloat64(mt.Data),
			Tags:                mt.Tags,
			NormalizerDataCache: p.NormalizerDataCache,
		}
		derivedMetric, err := p.DerivedMetrics.ProcessMetric(currentTime, metricData)
		if err != nil {
			return nil, errors.New("Unable to process metric: " + err.Error())
		}

		if derivedMetric != nil {
			namespaces := strings.Split(derivedMetric.Name, "/")
			newMetric := snap.Metric{
				Namespace: snap.NewNamespace(namespaces...),
				Version:   mt.Version,
				Tags:      mt.Tags,
				Timestamp: mt.Timestamp,
				Data:      derivedMetric.Value,
			}

			metrics = append(metrics, newMetric)
		}
	}

	return metrics, nil
}

// Analyze test analyze function
func (p *NodeAnalyzer) Analyze(mts []snap.Metric, cfg snap.Config) ([]snap.Metric, error) {
	if p.DerivedMetrics == nil {
		configs, ok := cfg["configs"]
		if !ok {
			return nil, snap.ErrConfigNotFound
		}

		dmCfgs := []DerivedMetricConfig{}
		for _, config := range configs.([]interface{}) {
			bytes, err := json.Marshal(config)
			if err != nil {
				return nil, fmt.Errorf("Unable to marshal src: %s", err)
			}

			dmCfg := DerivedMetricConfig{}
			err = json.Unmarshal(bytes, &dmCfg)
			if err != nil {
				return nil, fmt.Errorf("Unable to unmarshal into dst: %s", err)
			}

			if dmCfg.Normalizer != nil {
				p.NormalizerDataCache[*dmCfg.Normalizer] = 0
			}
			dmCfgs = append(dmCfgs, dmCfg)
		}

		interval, err := cfg.GetString("sampleInterval")
		if err != nil {
			return nil, errors.New("Unable to find sampleInterval duration: " + err.Error())
		}
		sampleInterval, err := strconv.ParseInt(interval, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse %s to int64: %s", interval, err.Error())
		}

		derivedMetrics, err := NewDerivedMetrics(sampleInterval, dmCfgs)
		if err != nil {
			return nil, errors.New("Unable to create derived metrics: " + err.Error())
		}

		p.DerivedMetrics = derivedMetrics
	}

	return p.ProcessMetrics(mts)
}

func convertFloat64(data interface{}) float64 {
	switch data.(type) {
	case int:
		return float64(data.(int))
	case int8:
		return float64(data.(int8))
	case int16:
		return float64(data.(int16))
	case int32:
		return float64(data.(int32))
	case int64:
		return float64(data.(int64))
	case uint64:
		return float64(data.(uint64))
	case float32:
		return float64(data.(float32))
	case float64:
		return float64(data.(float64))
	default:
		return float64(0)
	}
}
