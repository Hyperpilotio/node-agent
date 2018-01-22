package nodeanalyzer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/snap"
	log "github.com/sirupsen/logrus"
)

type NodeAnalyzer struct {
	DerivedMetrics *DerivedMetrics
	Normalizers    []string
}

type MetricConfigs struct {
	Configs []DerivedMetricConfig `json:"configs" binding:"required"`
}

func downloadConfigFile(url string) (*MetricConfigs, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, errors.New("Unable to download config file: " + err.Error())
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	configs := MetricConfigs{}
	if err := decoder.Decode(&configs); err != nil {
		return nil, errors.New("Unable to decode body: " + err.Error())
	}

	return &configs, nil
}

func init() {
	log.SetLevel(common.GetLevel(os.Getenv("SNAP_LOG_LEVEL")))
}

// NewProcessor generate processor
func NewAnalyzer() *NodeAnalyzer {
	return &NodeAnalyzer{
		Normalizers: make([]string, 0),
	}
}

func (p *NodeAnalyzer) ProcessMetrics(mts []snap.Metric) ([]snap.Metric, error) {
	metrics := []snap.Metric{}

	normalizersData := map[string]map[string]float64{}
	for _, mt := range mts {
		metricNm := strings.Join(mt.Namespace.Strings(), "/")
		for _, normalizersNm := range p.Normalizers {
			if metricNm == normalizersNm {
				normalizer := map[string]float64{}
				normalizer[mt.Tags["nodename"]] = convertFloat64(mt.Data)
				normalizersData[metricNm] = normalizer
			}
		}
	}

	for _, mt := range mts {
		currentTime := mt.Timestamp.UnixNano()
		metricData := &MetricData{
			MetricName:      strings.Join(mt.Namespace.Strings(), "/"),
			NodeName:        mt.Tags["nodename"],
			Value:           convertFloat64(mt.Data),
			Tags:            mt.Tags,
			NormalizersData: normalizersData,
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
		configUrl, err := cfg.GetString("configUrl")
		if err != nil {
			return nil, errors.New("Unable to find derived metrics config endpoint: " + err.Error())
		}

		configs, err := downloadConfigFile(configUrl)
		if err != nil {
			return nil, errors.New("Unable to download and deserialize configs: " + err.Error())
		}
		for _, dmCfg := range configs.Configs {
			if dmCfg.Normalizer != nil {
				p.Normalizers = append(p.Normalizers, *dmCfg.Normalizer)
			}
		}

		interval, err := cfg.GetString("sampleInterval")
		if err != nil {
			return nil, errors.New("Unable to find sampleInterval duration: " + err.Error())
		}
		sampleInterval, err := strconv.ParseInt(interval, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse %s to int64: %s", interval, err.Error())
		}

		derivedMetrics, err := NewDerivedMetrics(sampleInterval, configs.Configs)
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
