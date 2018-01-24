package nodeanalyzer

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"
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

type WindowStateHit struct {
	Hits         int64
	HitStartTime int64
	WindowTime   int64
	Count        int64
	TotalCount   int64
}

func (whs *WindowStateHit) addHits() {
	whs.Hits++
	if whs.Hits == 1 {
		whs.HitStartTime = time.Now().UnixNano()
	}
}

func (whs *WindowStateHit) addCount() {
	if whs.Hits > 0 {
		whs.Count++
	}
}

func (whs *WindowStateHit) getThresholdFrequency(currentTime int64) float64 {
	// There may be missing points when the time duration is over windowTime,
	// need to return the threshold frequency of this calculation
	hitDuration := currentTime - whs.HitStartTime
	if hitDuration >= whs.WindowTime {
		whs.Count = whs.TotalCount
	}

	if whs.Count == whs.TotalCount {
		thresholdFrequency := float64(whs.Hits) / float64(whs.TotalCount)
		whs.Hits = 0
		whs.Count = 0
		return thresholdFrequency
	}

	return -1
}

type ThresholdBasedState struct {
	MetricName          string
	DerivedMetricConfig *DerivedMetricConfig
	SampleInterval      int64
	WindowStateHit      *WindowStateHit
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
	if strings.Contains(metricName, "/*/") {
		metricName = strings.Replace(metricName, "/*/", "/", -1)
	}

	return &ThresholdBasedState{
		MetricName:          metricName,
		DerivedMetricConfig: config,
		SampleInterval:      sampleInterval,
		WindowStateHit: &WindowStateHit{
			WindowTime: config.ObservationWindowSec * 1000000000,
			TotalCount: totalCount,
		},
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
		tbs.WindowStateHit.addHits()
	}
	tbs.WindowStateHit.addCount()

	if value := tbs.WindowStateHit.getThresholdFrequency(currentTime); value != -1 {
		log.Infof("Finished compute %s threshold frequency: %f", tbs.MetricName, value)
		return &DerivedMetricResult{
			Name:  tbs.MetricName,
			Value: value,
		}
	}

	if tbs.WindowStateHit.Hits > 0 {
		log.Infof("%s[value:%f] threshold frequency[%s:%f] is %d/%d on latest time duration[%d/%d]",
			tbs.MetricName,
			metricValue,
			tbs.DerivedMetricConfig.ThresholdConfig.Type,
			tbs.DerivedMetricConfig.ThresholdConfig.Value,
			tbs.WindowStateHit.Hits,
			tbs.WindowStateHit.TotalCount,
			tbs.WindowStateHit.Count*tbs.SampleInterval/1000000000,
			tbs.DerivedMetricConfig.ObservationWindowSec,
		)
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
