package helpers

import (
	"log"
	"time"
)

type DumpFn func() error

var (
	defaultPeriod = time.Minute * 15
)

type Dumper struct {
	dumpFn    DumpFn
	period    time.Duration
	isUpdated chan struct{}
	stop      chan struct{}
}

func NewDumper(dumpFn DumpFn, period *time.Duration) *Dumper {
	if period == nil {
		period = &defaultPeriod
	}

	return &Dumper{
		dumpFn:    dumpFn,
		period:    *period,
		isUpdated: make(chan struct{}, 1),
		stop:      make(chan struct{}, 1),
	}
}

func (d *Dumper) Start() {
	go func() {
		for range time.Tick(d.period) {
			select {
			case <-d.isUpdated:
				d.dump()
			case <-d.stop:
				return
			default:
			}
		}
	}()
}

func (d *Dumper) ScheduleUpdate() {
	if len(d.isUpdated) == 0 {
		d.isUpdated <- struct{}{}
	}
}

func (d *Dumper) dump() {
	if err := d.dumpFn(); err != nil {
		log.Println("DUMPER: unexpected error: ", err.Error())
	}
}

func (d *Dumper) ScheduleStop() {
	d.stop <- struct{}{}
}
