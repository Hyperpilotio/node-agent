package nodeanalyzer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/snap"
	log "github.com/sirupsen/logrus"
)

type NodeAnalyzer struct {
	initialized       bool
	DerivedMetrics    *DerivedMetrics
	NormalizerMapping map[string]string
}

func init() {
	log.SetLevel(common.GetLevel(os.Getenv("SNAP_LOG_LEVEL")))
}

// NewProcessor generate processor
func NewAnalyzer() *NodeAnalyzer {
	return &NodeAnalyzer{
		NormalizerMapping: make(map[string]string),
	}
}

func (p *NodeAnalyzer) init(cfg snap.Config) error {
	configs, ok := cfg["configs"]
	if !ok {
		return snap.ErrConfigNotFound
	}

	dmCfgs := []DerivedMetricConfig{}
	for _, config := range configs.([]interface{}) {
		bytes, err := json.Marshal(config)
		if err != nil {
			return fmt.Errorf("Unable to marshal src: %s", err)
		}

		dmCfg := DerivedMetricConfig{}
		err = json.Unmarshal(bytes, &dmCfg)
		if err != nil {
			return fmt.Errorf("Unable to unmarshal into dst: %s", err)
		}

		if dmCfg.Normalizer != nil {
			p.NormalizerMapping[dmCfg.MetricName] = *dmCfg.Normalizer
		}
		dmCfgs = append(dmCfgs, dmCfg)
	}

	interval, err := cfg.GetString("sampleInterval")
	if err != nil {
		return errors.New("Unable to find sampleInterval duration: " + err.Error())
	}

	sampleInterval, err := time.ParseDuration(interval)
	if err != nil {
		return fmt.Errorf("Unable to parse %s to duration: %s", interval, err.Error())
	}

	derivedMetrics, err := NewDerivedMetrics(sampleInterval.Nanoseconds(), dmCfgs)
	if err != nil {
		return errors.New("Unable to create derived metrics: " + err.Error())
	}

	p.DerivedMetrics = derivedMetrics
	p.initialized = true

	return nil
}

func (p *NodeAnalyzer) replaceNamespaceWildcardValue(inputNm string, replaceNm string) string {
	actualNms := []string{}
	inputNms := strings.Split(inputNm, "/")
	replaceNms := strings.Split(replaceNm, "/")
	for i, element := range replaceNms {
		if element == "*" {
			element = inputNms[i]
		}
		actualNms = append(actualNms, element)
	}

	return strings.Join(actualNms, "/")
}

func (p *NodeAnalyzer) getNormalizerData(mts []snap.Metric) map[string]map[string]float64 {
	normalizerDataCache := map[string]map[string]float64{}
	for _, mt := range mts {
		mtMetricNm := "/" + strings.Join(mt.Namespace.Strings(), "/")
		mtNodename := mt.Tags["nodename"]
		for metricName, normalizer := range p.NormalizerMapping {
			if strings.Contains(metricName, "*") {
				metricName = p.replaceNamespaceWildcardValue(mtMetricNm, metricName)
				normalizer = p.replaceNamespaceWildcardValue(mtMetricNm, normalizer)
			}
			if mtMetricNm == normalizer {
				normalizerData := map[string]float64{}
				normalizerData[mtNodename] = convertFloat64(mt.Data)
				normalizerDataCache[metricName] = normalizerData
			}
		}
	}

	return normalizerDataCache
}

func (p *NodeAnalyzer) ProcessMetrics(mts []snap.Metric) ([]snap.Metric, error) {
	normalizerDataCache := p.getNormalizerData(mts)
	metrics := []snap.Metric{}
	for _, mt := range mts {
		metricNm := "/" + strings.Join(mt.Namespace.Strings(), "/")
		nodename := mt.Tags["nodename"]
		currentTime := mt.Timestamp.UnixNano()

		var normalizerData float64
		if data, ok := normalizerDataCache[metricNm]; ok {
			if val, ok := data[nodename]; ok {
				log.Infof("Find %s normalizer data %f", metricNm, val)
				normalizerData = val
			}
		}

		metricData := MetricData{
			MetricName:     metricNm,
			NodeName:       nodename,
			Value:          convertFloat64(mt.Data),
			Tags:           mt.Tags,
			NormalizerData: normalizerData,
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

		for _, globConfig := range p.DerivedMetrics.GlobConfigs {
			if globConfig.Pattern.Match(mtMetricNm) {
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

	if !p.initialized {
		if err := p.init(cfg); err != nil {
			return nil, err
		}
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
