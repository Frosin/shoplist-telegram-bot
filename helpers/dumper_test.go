package helpers_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/helpers"
	"github.com/magiconair/properties/assert"
	"go.uber.org/goleak"
)

var (
	testPeriod = time.Second
)

func TestDumper(t *testing.T) {
	var counter int32

	defer goleak.VerifyNone(t)

	dumpFn := func() error {
		atomic.AddInt32(&counter, 1)
		return nil
	}
	dumper := helpers.NewDumper(dumpFn, &testPeriod)
	dumper.Start()

	for i := 1; i < 21; i++ {
		dumper.ScheduleUpdate()
		time.Sleep(time.Millisecond * 100)
	}
	dumper.ScheduleStop()
	time.Sleep(testPeriod)

	assert.Equal(t, counter, int32(2))
}
