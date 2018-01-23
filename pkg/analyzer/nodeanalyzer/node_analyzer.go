package nodeanalyzer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/snap"
	log "github.com/sirupsen/logrus"
)

type NodeAnalyzer struct {
	DerivedMetrics    *DerivedMetrics
	MetricPatterns    []glob.Glob
	NormalizerMapping map[string]string
	mutex             sync.Mutex
}

func init() {
	log.SetLevel(common.GetLevel(os.Getenv("SNAP_LOG_LEVEL")))
}

// NewProcessor generate processor
func NewAnalyzer() *NodeAnalyzer {
	return &NodeAnalyzer{
		MetricPatterns:    make([]glob.Glob, 0),
		NormalizerMapping: make(map[string]string),
	}
}

func (p *NodeAnalyzer) getNormalizerData(filterMetricNm string, filterNodename string, mts []snap.Metric) float64 {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	filterNormalizerNm := ""
	for _, mt := range mts {
		mtMetricNm := "/" + strings.Join(mt.Namespace.Strings(), "/")
		mtNodename := mt.Tags["nodename"]
		dockerId := ""
		networkId := ""
		filesystemId := ""
		if strings.HasPrefix(mtMetricNm, "/intel/docker") {
			dockerId = mt.Namespace.Strings()[2]
			if strings.Contains(mtMetricNm, "/stats/network/") {
				networkId = mt.Namespace.Strings()[5]
			}
			if strings.Contains(mtMetricNm, "/stats/filesystem/") {
				filesystemId = mt.Namespace.Strings()[5]
			}
		}

		if mtMetricNm == filterMetricNm && mtNodename == filterNodename {
			for metricName, normalizer := range p.NormalizerMapping {
				if strings.HasPrefix(metricName, "/intel/docker") {
					metricName = strings.Replace(metricName, "/intel/docker/*/",
						fmt.Sprintf("/intel/docker/%s/", dockerId), 1)
					normalizer = strings.Replace(normalizer, "/intel/docker/*/",
						fmt.Sprintf("/intel/docker/%s/", dockerId), 1)

					metricName = strings.Replace(metricName, "/stats/network/*/",
						fmt.Sprintf("/stats/network/%s/", networkId), 1)
					normalizer = strings.Replace(normalizer, "/stats/network/*/",
						fmt.Sprintf("/stats/network/%s/", networkId), 1)

					metricName = strings.Replace(metricName, "/stats/filesystem/*/",
						fmt.Sprintf("/stats/filesystem/%s/", filesystemId), 1)
					normalizer = strings.Replace(normalizer, "/stats/filesystem/*/",
						fmt.Sprintf("/stats/filesystem/%s/", filesystemId), 1)
				}

				if mtMetricNm == metricName {
					filterNormalizerNm = normalizer
					break
				}
			}
		}
	}

	if filterNormalizerNm == "" {
		return 0
	}

	for _, mt := range mts {
		mtMetricNm := "/" + strings.Join(mt.Namespace.Strings(), "/")
		mtNodename := mt.Tags["nodename"]
		if mtMetricNm == filterNormalizerNm && mtNodename == filterNodename {
			normalizerValue := convertFloat64(mt.Data)
			log.Infof("Find %s normalizer data %f on %s", filterMetricNm, normalizerValue, filterNormalizerNm)
			return normalizerValue
		}
	}

	return 0
}

func (p *NodeAnalyzer) ProcessMetrics(mts []snap.Metric) ([]snap.Metric, error) {
	metrics := []snap.Metric{}
	for _, mt := range mts {
		metricNm := "/" + strings.Join(mt.Namespace.Strings(), "/")
		nodename := mt.Tags["nodename"]
		currentTime := mt.Timestamp.UnixNano()
		metricData := MetricData{
			MetricName:     metricNm,
			NodeName:       nodename,
			Value:          convertFloat64(mt.Data),
			Tags:           mt.Tags,
			NormalizerData: p.getNormalizerData(metricNm, nodename, mts),
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

func (p *NodeAnalyzer) getMetricTypes(mts []snap.Metric) ([]snap.Metric, error) {
	metrics := []snap.Metric{}
	for _, mt := range mts {
		mtMetricNm := strings.Join(mt.Namespace.Strings(), "/")
		if !strings.HasPrefix(mtMetricNm, "/") {
			mtMetricNm = "/" + mtMetricNm
		}

		for _, pattern := range p.MetricPatterns {
			if pattern.Match(mtMetricNm) {
				metrics = append(metrics, mt)
				break
			}
		}
	}

	return metrics, nil
}

func (p *NodeAnalyzer) Analyze(mts []snap.Metric, cfg snap.Config) ([]snap.Metric, error) {
	if len(mts) == 0 {
		return nil, errors.New("Unable to get metrics to analyze")
	}

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
				p.NormalizerMapping[dmCfg.MetricName] = *dmCfg.Normalizer
			}
			dmCfgs = append(dmCfgs, dmCfg)

			pattern, err := glob.Compile(dmCfg.MetricName)
			if err != nil {
				return nil, fmt.Errorf("Unable to compile pattern %s: %s", dmCfg.MetricName, err.Error())
			}
			p.MetricPatterns = append(p.MetricPatterns, pattern)
		}

		interval, err := cfg.GetString("sampleInterval")
		if err != nil {
			return nil, errors.New("Unable to find sampleInterval duration: " + err.Error())
		}

		sampleInterval, err := time.ParseDuration(interval)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse %s to duration: %s", interval, err.Error())
		}

		derivedMetrics, err := NewDerivedMetrics(sampleInterval.Nanoseconds(), dmCfgs)
		if err != nil {
			return nil, errors.New("Unable to create derived metrics: " + err.Error())
		}

		p.DerivedMetrics = derivedMetrics
	}

	analyzeMts, err := p.getMetricTypes(mts)
	if err != nil {
		return nil, errors.New("Unable to get metric types to analyze: " + err.Error())
	}

	if len(analyzeMts) == 0 {
		return nil, fmt.Errorf("No metrics are needed to analyze")
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
