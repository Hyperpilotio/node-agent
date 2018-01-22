package nodeanalyzer

import (
	"errors"
)

type DerivedMetricCalculator interface {
	GetDerivedMetric(currentTime int64, metricData *MetricData) *DerivedMetricResult
}

type MetricData struct {
	MetricName      string
	Value           float64
	NodeName        string
	Tags            map[string]string
	NormalizersData map[string]map[string]float64
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
	Hits                []int64
	TotalCount          int64
	SampleInterval      int64
}

func NewThresholdBasedState(sampleInterval int64, config *DerivedMetricConfig) *ThresholdBasedState {
	totalCount := config.ObservationWindowSec / sampleInterval
	metricName := config.MetricName + "/" + config.Type
	if config.Normalizer != nil {
		metricName = config.MetricName + "_normalized/" + config.Type
	}

	return &ThresholdBasedState{
		MetricName:          metricName,
		DerivedMetricConfig: config,
		Hits:                make([]int64, 0),
		TotalCount:          totalCount,
		SampleInterval:      sampleInterval,
	}
}

func NewDerivedMetricCalculator(sampleInterval int64, config *DerivedMetricConfig) (DerivedMetricCalculator, error) {
	if config.ThresholdConfig != nil {
		return NewThresholdBasedState(sampleInterval, config), nil
	}

	return nil, errors.New("No metric config found")
}

func (tbs *ThresholdBasedState) matchFilterTags(metricData *MetricData) bool {
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

	if tbs.DerivedMetricConfig.Type == "UB" {
		return value >= tbs.DerivedMetricConfig.ThresholdConfig.Value
	} else {
		return value <= tbs.DerivedMetricConfig.ThresholdConfig.Value
	}
}

func (tbs *ThresholdBasedState) GetDerivedMetric(currentTime int64, metricData *MetricData) *DerivedMetricResult {
	if !tbs.matchFilterTags(metricData) {
		return nil
	}

	var metricValue float64
	if tbs.DerivedMetricConfig.Normalizer != nil {
		normalizerValue := metricData.NormalizersData[*tbs.DerivedMetricConfig.Normalizer][metricData.NodeName]
		metricValue = (metricData.Value / normalizerValue)
	} else {
		metricValue = metricData.Value
	}

	if tbs.computeSeverity(metricValue) {
		hitsLength := len(tbs.Hits)
		if hitsLength > 0 {
			// Fill in missing metric points
			for currentTime-tbs.Hits[len(tbs.Hits)-1] >= 2*tbs.SampleInterval {
				filledHitTime := tbs.Hits[len(tbs.Hits)-1] + tbs.SampleInterval
				tbs.Hits = append(tbs.Hits, filledHitTime)
			}
		}

		// Prune values outside of window
		windowBeginTime := currentTime - tbs.DerivedMetricConfig.ObservationWindowSec
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

	return &DerivedMetricResult{
		Name:  tbs.MetricName,
		Value: float64(len(tbs.Hits)) / float64(tbs.TotalCount),
	}
}

type DerivedMetrics struct {
	States map[string]DerivedMetricCalculator
}

func NewDerivedMetrics(sampleInterval int64, configs []DerivedMetricConfig) (*DerivedMetrics, error) {
	states := map[string]DerivedMetricCalculator{}
	for _, config := range configs {
		calculator, err := NewDerivedMetricCalculator(sampleInterval, &config)
		if err != nil {
			return nil, errors.New("Unable to create derived metric calculator: " + err.Error())
		}
		states[config.MetricName] = calculator
	}

	return &DerivedMetrics{
		States: states,
	}, nil
}

func (dm *DerivedMetrics) ProcessMetric(currentTime int64, metricData *MetricData) (*DerivedMetricResult, error) {
	state, ok := dm.States[metricData.MetricName]
	if !ok {
		return nil, nil
	}

	return state.GetDerivedMetric(currentTime, metricData), nil
}
