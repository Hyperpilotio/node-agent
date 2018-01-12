package prometheus

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
	pluginName      = "prometheus"
	pluginVersion   = int64(1)
	namespacePrefix = []string{vendor, pluginName}
	configFile      = "/etc/snap-configs/snap-plugin-collector-prometheus-config"
)

var prometheusEndpoint string = "http://localhost:8080/metrics"

type MetricsDownloader interface {
	GetMetricsReader(url string) (io.Reader, error)
	GetEndpoint(config snap.Config) (string, error)
}

type HTTPMetricsDownloader struct {
}

// PrometheusCollector struct
type PrometheusCollector struct {
	Downloader MetricsDownloader
}

func init() {
	log.SetLevel(common.GetLevel(os.Getenv("SNAP_LOG_LEVEL")))
}

// New return an instance of PrometheusCollector
func New() (*PrometheusCollector, error) {
	return &PrometheusCollector{
		Downloader: HTTPMetricsDownloader{},
	}, nil
}

func createMetricFromFamily(currentTime time.Time, metricFamily *dto.MetricFamily) snap.Metric {
	fullNamespace := namespacePrefix
	fullNamespace = append(fullNamespace, *metricFamily.Name)
	return snap.Metric{
		Namespace:   snap.NewNamespace(fullNamespace...),
		Timestamp:   currentTime,
		Description: metricFamily.GetHelp(),
		Version:     pluginVersion,
	}
}

func (c *PrometheusCollector) _collectMetrics(mts []snap.Metric) ([]snap.Metric, error) {
	var err error
	var metrics []snap.Metric
	currentTime := time.Now()

	if len(mts) == 0 {
		return metrics, fmt.Errorf("array of metric type is empty\nPlease check GetMetricTypes()")
	}

	endpoint, err := c.Downloader.GetEndpoint(mts[0].Config)
	if err != nil {
		return metrics, fmt.Errorf("Unable to get endpoint: " + err.Error())
	}

	metricFamilies, err := c.Collect(endpoint)
	if err != nil {
		log.Warningf("Unable to collect metrics, skipping to next cycle. endpoint: %s, error: %s", endpoint, err.Error())
		return metrics, nil
	}

	for _, metricFamily := range metricFamilies {
		for _, metricItem := range metricFamily.GetMetric() {
			switch metricFamily.GetType() {
			case dto.MetricType_GAUGE:
				metric := createMetricFromFamily(currentTime, metricFamily)
				if strings.Contains(metricFamily.GetName(), "bytes") {
					metric.Unit = "B"
				}
				metric.Data = metricItem.GetGauge().GetValue()
				metric.Tags = getTagsOfMetric(metricItem)
				metrics = append(metrics, metric)

			case dto.MetricType_COUNTER:
				metric := createMetricFromFamily(currentTime, metricFamily)
				metric.Data = metricItem.GetCounter().GetValue()
				metric.Tags = getTagsOfMetric(metricItem)
				metrics = append(metrics, metric)

			case dto.MetricType_SUMMARY:
				summaryData, err := processSummaryMetric(metricItem)
				if err != nil {
					continue
				}
				for key, val := range summaryData {
					metric := createMetricFromFamily(currentTime, metricFamily)
					tags := getTagsOfMetric(metricItem)
					tags["summary"] = key
					metric.Tags = tags
					metric.Data = val
					metrics = append(metrics, metric)
				}
			}
		}
	}

	return metrics, nil
}

// CollectMetrics will be called by Snap when a task that collects one of the metrics returned from this plugins
func (c *PrometheusCollector) CollectMetrics(mts []snap.Metric) ([]snap.Metric, error) {
	var (
		metrics []snap.Metric
		err     error
	)

	metrics, err = c._collectMetrics(mts)
	if err != nil {
		return mts, err
	}

	return metrics, nil
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

func (c PrometheusCollector) Collect(endpoint string) (map[string]*dto.MetricFamily, error) {
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

func (c *PrometheusCollector) GetMetricTypes(cfg snap.Config) ([]snap.Metric, error) {
	mts := []snap.Metric{}
	mts = append(mts, snap.Metric{
		Namespace: snap.NewNamespace(namespacePrefix...),
		Version:   pluginVersion,
	})

	return mts, nil
}
