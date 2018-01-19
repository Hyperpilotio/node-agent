package nodeanalyzer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gobwas/glob"
)

type DerivedMetric struct {
	Name  string
	Value float64
}

type DerivedMetricCalculator interface {
	GetDerivedMetric(currentTime int64, value float64) *DerivedMetric
}

type DerivedMetricConfig struct {
	MetricName           string            `json:"metric_name"`
	Type                 string            `json:"type"`
	Resource             string            `json:"resource"`
	Normalizer           string            `json:"normalizer,omitempty"`
	ObservationWindowSec int64             `json:"observation_window_sec"`
	Tags                 map[string]string `json:"tags"`

	ThresholdBased *ThresholdBasedConfig `json:"threshold"`
}

type ThresholdBasedConfig struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
}

type ThresholdBasedState struct {
	Config         *DerivedMetricConfig
	Hits           []int64
	TotalCount     int64
	SampleInterval int64
}

func NewThresholdBasedState(sampleInterval int64, config *DerivedMetricConfig) *ThresholdBasedState {
	totalCount := config.ObservationWindowSec / sampleInterval
	return &ThresholdBasedState{
		Config:         config,
		Hits:           make([]int64, 0),
		TotalCount:     totalCount,
		SampleInterval: sampleInterval,
	}
}

func NewDerivedMetricCalculator(sampleInterval int64, config *DerivedMetricConfig) (DerivedMetricCalculator, error) {
	if config.ThresholdBased != nil {
		return NewThresholdBasedState(sampleInterval, config), nil
	}

	return nil, errors.New("No metric config found")
}

func (tbs *ThresholdBasedState) isMatchThresholdType(value float64) bool {
	if tbs.Config.Type == "UB" {
		return value >= tbs.Config.ThresholdBased.Value
	} else {
		return value <= tbs.Config.ThresholdBased.Value
	}
}

func (tbs *ThresholdBasedState) GetDerivedMetric(currentTime int64, value float64) *DerivedMetric {
	if tbs.isMatchThresholdType(value) {
		hitsLength := len(tbs.Hits)
		if hitsLength > 0 {
			// Fill in missing metric points
			for currentTime-tbs.Hits[len(tbs.Hits)-1] >= 2*tbs.SampleInterval {
				filledHitTime := tbs.Hits[len(tbs.Hits)-1] + tbs.SampleInterval
				tbs.Hits = append(tbs.Hits, filledHitTime)
			}
		}

		// Prune values outside of window
		windowBeginTime := currentTime - tbs.Config.ObservationWindowSec
		lastGoodIndex := -1
		for i, hit := range tbs.Hits {
			if hit >= windowBeginTime {
				lastGoodIndex = i
				break
			}
		}

		if lastGoodIndex == -1 {
			// All values are outside of window, clear all values
			tbs.Hits = []int64{}
		} else {
			tbs.Hits = tbs.Hits[lastGoodIndex:]
		}

		tbs.Hits = append(tbs.Hits, currentTime)
	}

	return &DerivedMetric{
		Name:  tbs.Config.MetricName,
		Value: float64(len(tbs.Hits)) / float64(tbs.TotalCount),
	}
}

type GlobConfig struct {
	Config  *DerivedMetricConfig
	Pattern glob.Glob
}

type DerivedMetrics struct {
	States         map[string]DerivedMetricCalculator
	SampleInterval int64
	GlobConfigs    []GlobConfig
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
		SampleInterval: sampleInterval,
		GlobConfigs:    globConfigs,
	}, nil
}

func (ae *DerivedMetrics) ProcessMetric(currentTime int64, metricName string, value float64) (*DerivedMetric, error) {
	state, ok := ae.States[metricName]
	if !ok {
		for _, globConfig := range ae.GlobConfigs {
			if globConfig.Pattern.Match(metricName) {
				calculator, err := NewDerivedMetricCalculator(ae.SampleInterval, globConfig.Config)
				if err != nil {
					return nil, fmt.Errorf("Unable to create state for metric %s: %s", metricName, err.Error())
				}
				ae.States[metricName] = calculator
				state = calculator
				break
			}
		}

		if state == nil {
			return nil, nil
		}
	}

	return state.GetDerivedMetric(currentTime, value), nil
}
