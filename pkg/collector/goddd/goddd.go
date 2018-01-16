package goddd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/snap"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	log "github.com/sirupsen/logrus"
)

var (
	vendor          = "hyperpilot"
	pluginName      = "goddd"
	pluginVersion   = 1
	nameSpacePrefix = []string{vendor, pluginName}
)

type MetricsDownloader interface {
	GetMetricsReader(url string) (io.Reader, error)
	GetEndpoint(config snap.Config) (string, error)
}

type HTTPMetricsDownloader struct {
}

// GodddCollector struct
type GodddCollector struct {
	Downloader MetricsDownloader
	cache      *CacheType
}

func init() {
	log.SetLevel(common.GetLevel(os.Getenv("SNAP_LOG_LEVEL")))
}

// New return an instance of Goddd
func New() (*GodddCollector, error) {
	return &GodddCollector{
		Downloader: HTTPMetricsDownloader{},
		cache:      NewCache(),
	}, nil
}

func (c *GodddCollector) _collectMetrics(mts []snap.Metric) ([]metricWithType, error) {
	var err error
	metrics := []metricWithType{}
	currentTime := time.Now()

	if len(mts) == 0 {
		return metrics, fmt.Errorf("array of metric type is empty\nPlease check GetMetricTypes()")
	}

	endpoint, err := c.Downloader.GetEndpoint(mts[0].Config)
	if err != nil {
		return metrics, fmt.Errorf("Unable to get endpoint: " + err.Error())
	}

	metricFamilies, err := c.collect(endpoint)
	if err != nil {
		log.Warningf("Unable to collect metrics, skipping to next cycle. endpoint: %s, error: %s", endpoint, err.Error())
		return metrics, nil
	}

	for idx, mt := range mts {
		mts[idx].Timestamp = currentTime
		ns := mt.Namespace.Strings()
		metricFamily := metricFamilies[ns[len(ns)-1]]

		metric := metricWithType{
			Type: metricFamily.GetType(),
		}
		metric.Namespace = snap.NewNamespace(ns...)
		metric.Timestamp = currentTime
		metric.Description = metricFamily.GetHelp()
		metric.Version = int64(pluginVersion)

		metricsBuffer := []metricWithType{}

		for _, metricOfGoddd := range metricFamily.GetMetric() {
			switch metricFamily.GetType() {

			case dto.MetricType_GAUGE:
				if strings.Contains(metricFamily.GetName(), "bytes") {
					metric.Unit = "B"
				}
				metric.Data = metricOfGoddd.GetGauge().GetValue()
				metric.Tags = getTagsOfMetric(metricOfGoddd)
				metricsBuffer = append(metricsBuffer, metric)

			case dto.MetricType_COUNTER:
				metric.Data = metricOfGoddd.GetCounter().GetValue()
				metric.Tags = getTagsOfMetric(metricOfGoddd)
				metricsBuffer = append(metricsBuffer, metric)

			case dto.MetricType_SUMMARY:
				summaryData, err := processSummaryMetric(metricOfGoddd)
				if err != nil {
					continue
				}
				for key, val := range summaryData {
					tags := getTagsOfMetric(metricOfGoddd)
					tags["summary"] = key
					metric.Tags = tags
					metric.Data = val
					metricsBuffer = append(metricsBuffer, metric)
				}

			}
		}

		for _, val := range MultiGroupsMetricList {
			if val == metricFamily.GetName() {
				switch metricFamily.GetType() {

				case dto.MetricType_COUNTER:
					var totalCount float64
					for _, collectedMetric := range metricsBuffer {
						totalCount += collectedMetric.Data.(float64)
					}
					metric.Data = totalCount
					metric.Tags = map[string]string{"total": "TOTAL"}
					metricsBuffer = append(metricsBuffer, metric)

				case dto.MetricType_SUMMARY:
					var totalCount, totalSum float64
					for _, collectedMetric := range metricsBuffer {
						switch collectedMetric.Tags["summary"] {
						case "count":
							totalCount += collectedMetric.Data.(float64)
						case "sum":
							totalSum += collectedMetric.Data.(float64)
						}
					}
					metric.Data = totalCount
					metric.Tags = map[string]string{
						"total":   "TOTAL",
						"summary": "count",
					}
					metricsBuffer = append(metricsBuffer, metric)
					metric.Data = totalSum
					metric.Tags = map[string]string{
						"total":   "TOTAL",
						"summary": "sum",
					}
					metricsBuffer = append(metricsBuffer, metric)
					metric.Data = totalSum / totalCount
					metric.Tags = map[string]string{
						"total":   "TOTAL",
						"summary": "avg",
					}
					metricsBuffer = append(metricsBuffer, metric)
				}
			}
		}

		for _, collectedMetric := range metricsBuffer {
			metrics = append(metrics, collectedMetric)
		}
	}
	return metrics, nil
}

// CollectMetrics will be called by Snap when a task that collects one of the metrics returned from this plugins
func (c *GodddCollector) CollectMetrics(mts []snap.Metric) ([]snap.Metric, error) {
	var (
		metrics []metricWithType
		res     []snap.Metric
		err     error
	)

	metrics, err = c._collectMetrics(mts)
	if err != nil {
		return mts, err
	}

	res, err = c._cache(metrics)
	if err != nil {
		return mts, err
	}

	return res, nil
}

func getTagsOfMetric(metric *dto.Metric) map[string]string {
	tags := make(map[string]string)
	for _, label := range metric.GetLabel() {
		tags[label.GetName()] = label.GetValue()
	}
	return tags
}

func processSummaryMetric(metric *dto.Metric) (map[string]float64, error) {
	summary := make(map[string]float64)
	summary["count"] = float64(metric.GetSummary().GetSampleCount())
	summary["sum"] = float64(metric.GetSummary().GetSampleSum())
	summary["avg"] = summary["sum"] / summary["count"]

	for _, quantile := range metric.GetSummary().GetQuantile() {
		key := fmt.Sprintf("quantile_%d", int(quantile.GetQuantile()*100))
		if !math.IsNaN(quantile.GetValue()) {
			summary[key] = quantile.GetValue()
		} else {
			log.Warningf("Skipping to write metric %s as it's value is NaN", key)
		}
	}

	return summary, nil
}

func (downloader HTTPMetricsDownloader) GetEndpoint(config snap.Config) (string, error) {
	address, err := config.GetString("endpoint")
	if err != nil {
		return "", err
	}

	if strings.Contains(address, "/metrics") {
		return address, nil
	}

	return address + "/metrics", nil
}

func (downloader HTTPMetricsDownloader) GetMetricsReader(url string) (io.Reader, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return nil, err
	} else if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()

		// Copy content from the body of http request
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		b := buf.Bytes()
		httpBody := bytes.NewReader(b)

		return httpBody, nil
	} else {
		return nil, fmt.Errorf("Status code: %d Response: %v\n", resp.StatusCode, resp)
	}
}

func parseMetrics(httpBody io.Reader) (map[string]*dto.MetricFamily, error) {
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(httpBody)
	if err != nil {
		fmt.Println(err)
		return make(map[string]*dto.MetricFamily), err
	}
	return metricFamilies, nil
}

func (c GodddCollector) collect(endpoint string) (map[string]*dto.MetricFamily, error) {
	reader, err := c.Downloader.GetMetricsReader(endpoint)
	if err != nil {
		return nil, errors.New("Unable to download metrics: " + err.Error())
	}
	metricFamilies, err := parseMetrics(reader)
	if err != nil {
		return nil, errors.New("Unable to parse metrics: " + err.Error())
	}
	return metricFamilies, nil
}

//GetMetricTypes returns metric types for testing
func (c *GodddCollector) GetMetricTypes(cfg snap.Config) ([]snap.Metric, error) {
	mts := []snap.Metric{}

	for _, val := range MetricList {
		mts = append(mts, snap.Metric{
			// /hyperpilot/goddd/*
			Namespace: snap.NewNamespace(nameSpacePrefix...).
				AddStaticElement(val),
			Version: int64(pluginVersion),
		})
	}

	return mts, nil
}
