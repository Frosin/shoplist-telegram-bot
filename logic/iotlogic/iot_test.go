package iotlogic

import (
	"testing"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/iot"
	"github.com/magiconair/properties/assert"
)

func Test_iot(t *testing.T) {
	for _, v := range []struct {
		t time.Time
		e float64
	}{
		{
			t: createTime("15:30:44"),
			e: 15.5,
		},
		{
			t: createTime("20:00:44"),
			e: 20.0,
		},
		{
			t: createTime("20:14:44"),
			e: 20.23,
		},
		{
			t: createTime("20:15:44"),
			e: 20.25,
		},
	} {

		real := timeToFloat(v.t)
		assert.Equal(t, real, v.e)
	}
}

func createTime(s string) time.Time {
	t, _ := time.Parse(iot.TimeLayout, s)
	return t
}
