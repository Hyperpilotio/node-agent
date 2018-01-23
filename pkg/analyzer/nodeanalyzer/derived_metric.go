package nodeanalyzer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gobwas/glob"
)

type DerivedMetricCalculator interface {
	GetDerivedMetric(currentTime int64, metricData MetricData) *DerivedMetricResult
}

type MetricData struct {
	MetricName     string
	Value          float64
	NodeName       string
	Tags           map[string]string
	NormalizerData float64
}

type DerivedMetricResult struct {
	Name  string
	Value float64
}

type DerivedMetricConfig struct {
	MetricName           string            `json:"metric_name"`
	Type                 string            `json:"type"`
	Resource             string            `json:"resource"`
	Normalizer           *string           `json:"normalizer,omitempty"`
	ObservationWindowSec int64             `json:"observation_window_sec"`
	Tags                 map[string]string `json:"tags,omitempty"`

	ThresholdConfig *ThresholdBasedConfig `json:"threshold"`
}

type ThresholdBasedConfig struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type ThresholdBasedState struct {
	MetricName          string
	DerivedMetricConfig *DerivedMetricConfig
	Hits                int64
	Count               int64
	TotalCount          int64
}

func NewThresholdBasedState(sampleInterval int64, config *DerivedMetricConfig) *ThresholdBasedState {
	totalCount := config.ObservationWindowSec * 1000000000 / sampleInterval
	metricName := config.MetricName + "/" + config.Type
	if config.Normalizer != nil {
		metricName = config.MetricName + "_normalized/" + config.Type
	}
	if strings.HasPrefix(metricName, "/") {
		metricName = metricName[1:]
	}

	return &ThresholdBasedState{
		MetricName:          metricName,
		DerivedMetricConfig: config,
		TotalCount:          totalCount,
	}
}

func NewDerivedMetricCalculator(sampleInterval int64, config *DerivedMetricConfig) (DerivedMetricCalculator, error) {
	if config.ThresholdConfig != nil {
		return NewThresholdBasedState(sampleInterval, config), nil
	}

	return nil, errors.New("No metric config found")
}

func (tbs *ThresholdBasedState) matchFilterTags(metricData MetricData) bool {
	if tbs.DerivedMetricConfig.Tags != nil {
		for k, v := range tbs.DerivedMetricConfig.Tags {
			if v != metricData.Tags[k] {
				return false
			}
		}
	}
	return true
}

func (tbs *ThresholdBasedState) computeSeverity(value float64) bool {
	if tbs.DerivedMetricConfig.ThresholdConfig.Unit == "ms" {
		value = value / 1000
	}

	if tbs.DerivedMetricConfig.ThresholdConfig.Type == "UB" {
		return value >= tbs.DerivedMetricConfig.ThresholdConfig.Value
	} else {
		return value <= tbs.DerivedMetricConfig.ThresholdConfig.Value
	}
}

func (tbs *ThresholdBasedState) GetDerivedMetric(currentTime int64, metricData MetricData) *DerivedMetricResult {
	if !tbs.matchFilterTags(metricData) {
		return nil
	}

	var metricValue float64
	if tbs.DerivedMetricConfig.Normalizer != nil {
		metricValue = (metricData.Value / metricData.NormalizerData)
	} else {
		metricValue = metricData.Value
	}

	if tbs.computeSeverity(metricValue) {
		tbs.Hits++
	}

	if tbs.Hits > 0 {
		tbs.Count++
	}

	if tbs.Count == tbs.TotalCount {
		value := float64(tbs.Hits) / float64(tbs.TotalCount)
		tbs.Hits = 0
		tbs.Count = 0

		return &DerivedMetricResult{
			Name:  tbs.MetricName,
			Value: value,
		}
	}

	return nil
}

type GlobConfig struct {
	Config  *DerivedMetricConfig
	Pattern glob.Glob
}

type DerivedMetrics struct {
	States         map[string]DerivedMetricCalculator
	GlobConfigs    []GlobConfig
	SampleInterval int64
}

func NewDerivedMetrics(sampleInterval int64, configs []DerivedMetricConfig) (*DerivedMetrics, error) {
	states := map[string]DerivedMetricCalculator{}
	globConfigs := []GlobConfig{}

	for _, config := range configs {
		// We assume any metric name with wildcard is a pattern to be matched
		if strings.Contains(config.MetricName, "/*") {
			pattern, err := glob.Compile(config.MetricName)
			if err != nil {
				return nil, fmt.Errorf("Unable to compile pattern %s: %s", config.MetricName, err.Error())
			}
			globConfig := GlobConfig{
				Config:  &config,
				Pattern: pattern,
			}
			globConfigs = append(globConfigs, globConfig)
		} else {
			calculator, err := NewDerivedMetricCalculator(sampleInterval, &config)
			if err != nil {
				return nil, errors.New("Unable to create derived metric calculator: " + err.Error())
			}
			states[config.MetricName] = calculator
		}
	}

	return &DerivedMetrics{
		States:         states,
		GlobConfigs:    globConfigs,
		SampleInterval: sampleInterval,
	}, nil
}

func (dm *DerivedMetrics) ProcessMetric(currentTime int64, metricData MetricData) (*DerivedMetricResult, error) {
	state, ok := dm.States[metricData.MetricName]
	if !ok {
		for _, globConfig := range dm.GlobConfigs {
			if globConfig.Pattern.Match(metricData.MetricName) {
				calculator, err := NewDerivedMetricCalculator(dm.SampleInterval, globConfig.Config)
				if err != nil {
					return nil, fmt.Errorf("Unable to create state for metric %s: %s", metricData.MetricName, err.Error())
				}
				dm.States[metricData.MetricName] = calculator
				state = calculator
				break
			}
		}

		if state == nil {
			return nil, nil
		}
	}

	return state.GetDerivedMetric(currentTime, metricData), nil
}
