package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStartMetricsUpdater(t *testing.T) {
	strg := NewMetricStorage()
	testChan1 := strg.AddMetric("test1", "test1", "test1")
	testChan2 := strg.AddMetric("test2", "test2", "test2")

	testChan1 <- 5.55
	testChan2 <- 7.77

	go strg.StartMetricsUpdater(time.Millisecond * 100)

	time.Sleep(time.Millisecond * 110)

	expected := []float64{5.55, 7.77}
	for i, v := range strg.metrics {
		assert.Equal(t, expected[i], v.current)
	}
}
