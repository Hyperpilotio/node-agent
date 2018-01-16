package agent

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"

	"github.com/go-resty/resty"
	"github.com/hyperpilotio/node-agent/pkg/common"
	"github.com/hyperpilotio/node-agent/pkg/snap"
	log "github.com/sirupsen/logrus"
)

// Processor test processor
type GodddQoSProcessor struct {
	Goal float64
}

func init() {
	log.SetLevel(common.GetLevel(os.Getenv("SNAP_LOG_LEVEL")))
}

// NewProcessor generate processor
func NewProcessor() *GodddQoSProcessor {
	return &GodddQoSProcessor{}
}

func SendMetricToQosDataStore(urlString string, slack float64, latency float64) error {
	u, err := url.Parse(urlString)
	if err != nil {
		return errors.New("Unable to parse url: " + err.Error())
	}

	u.Path = path.Join(u.Path, "v1/apps/goddd/metrics")
	body := make(map[string]float64)
	body["slack"] = slack
	body["latency"] = latency
	if response, err := resty.R().SetBody(body).Post(u.String()); err != nil {
		return errors.New("Unable to send metrics to qos data store: " + err.Error())
	} else if response.StatusCode() >= 300 {
		return fmt.Errorf("Unable to send metrics to qos data store: Unexpected response code %d, body: %s", response.StatusCode(), response.String())
	}

	return nil
}

func calculateSlackValue(goal float64, current float64) (float64, error) {
	if goal == 0 {
		return goal, errors.New("SLA goal value should not be zero!")
	}

	return (goal - current) / goal, nil
}

// Process test process function
func (p *GodddQoSProcessor) Process(mts []snap.Metric, cfg snap.Config) ([]snap.Metric, error) {
	goalConfig, err := cfg.GetInt("sla-goal")
	if err != nil {
		return mts, errors.New("Unable to read sla-goal config: " + err.Error())
	}
	slaGoal := float64(goalConfig) / 1000.0 // convert slo value from milliseconds to seconds

	metricType, err := cfg.GetString("metric-type")
	if err != nil {
		return mts, errors.New("Unable to read metric-type config: " + err.Error())
	}

	fmt.Println("SLA goal for snap-goddd processor: " + strconv.FormatFloat(slaGoal, 'f', -1, 64) + " seconds")
	fmt.Println("Metric type for snap-goddd processor: " + metricType)

	var metricSum float64
	var count int
	for _, mt := range mts {
		if mt.Namespace.Strings()[len(mt.Namespace.Strings())-1] == "api_booking_service_request_latency_microseconds" { // metric value in seconds
			if summary, ok := mt.Tags["summary"]; ok {
				if summary == metricType {
					metricSum += mt.Data.(float64)
					count++
				}
			}
		}
	}

	if count > 0 {
		metricMean := metricSum / float64(count)
		slackValue, err := calculateSlackValue(slaGoal, metricMean)
		if err != nil {
			return mts, errors.New("Unable to calculate slack value: " + err.Error())
		}

		urlString, err := cfg.GetString("qos-data-store-url")
		if err != nil {
			return mts, errors.New("Unable to read qos-data-store-url config: " + err.Error())
		}

		if err := SendMetricToQosDataStore(urlString, slackValue, metricMean); err != nil {
			return mts, errors.New("Unable to send slack value: " + err.Error())
		}
	}

	return mts, nil
}
