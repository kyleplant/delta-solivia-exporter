package exporter

import (
	"math/rand"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "delta-solivia"
)

var (
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last query of Delta Solivia inverter successful.",
		nil, nil,
	)
	acPower = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "ac_power"),
		"The ac power reading.",
		nil, nil,
	)
)

// Exporter collects power stats from the given serial device connected to a
// Delta Solivia inverter and exports them using the prometheus metrics package.
type Exporter struct {
	healthSummary bool
	logger        log.Logger
}

// SerialOpts configures options for connecting to serial device connected to inverter.
type SerialOpts struct {
	Address      string
	BaudRate     int
	Timeout      time.Duration
	Insecure     bool
	RequestLimit int
}

// New returns an initialized Exporter.
func New(opts SerialOpts, healthSummary bool, logger log.Logger) (*Exporter, error) {

	// Init our exporter.
	return &Exporter{
		healthSummary: healthSummary,
		logger:        logger,
	}, nil
}

// Describe describes all the metrics ever exported by the Delta Solivia
// exporter. It implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- acPower
}

// Collect fetches the stats from configured inverter and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	ok := e.collectAcPower(ch)

	if ok {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 1.0,
		)
	} else {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0.0,
		)
	}
}

func (e *Exporter) collectAcPower(ch chan<- prometheus.Metric) bool {
	rand.Seed(time.Now().UnixNano())
	min := 1000
	max := 1200
	ch <- prometheus.MustNewConstMetric(
		acPower, prometheus.GaugeValue, float64(rand.Intn(max-min+1)+min),
	)
	return true
}
