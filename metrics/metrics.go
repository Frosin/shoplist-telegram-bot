package metrics

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metric struct {
	name       string
	gauge      prometheus.Gauge
	current    float64
	sourceChan chan float64
}

type MetricStorage struct {
	metrics []metric
}

type UpdateResult struct {
	QueueName string
	Err       error
	Len       float64
}

//NewMetricStorage create metrics object from config data
func NewMetricStorage() *MetricStorage {
	metricStorage := MetricStorage{}

	return &metricStorage
}

func (m *MetricStorage) AddMetric(name, namespace, subsystem string) chan float64 {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      strings.ReplaceAll(name, ".", "_"), // metric name must not contain '.' symbol
		Help:      fmt.Sprintf("shoplist internal metric '%s'", name),
	})
	metric := metric{
		name:       name,
		gauge:      gauge,
		sourceChan: make(chan float64, 1),
	}
	m.metrics = append(m.metrics, metric)

	return metric.sourceChan
}

//GetMetricsHandler returns default prometheus client server handler
func (m *MetricStorage) GetMetricsHandler() http.Handler {
	r := prometheus.NewRegistry()

	for _, metric := range m.metrics {
		r.MustRegister(metric.gauge)
	}

	return promhttp.HandlerFor(r, promhttp.HandlerOpts{})
}

//StartMetricsUpdater returns metric updater job
func (m *MetricStorage) StartMetricsUpdater(updateInterval time.Duration) {
	for range time.Tick(updateInterval) {
		for i, metric := range m.metrics {
			currentValue := metric.current
			// get actual value
			select {
			case value := <-metric.sourceChan:
				currentValue = value
			default:
			}
			m.metrics[i].current = currentValue
			metric.gauge.Set(currentValue)
		}
	}
}
