package helpers

import (
	"log"
	"time"
)

type DumpFn func() error

var (
	defaultPeriod = time.Minute * 1
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
			// debug
			log.Println("in tick")
			select {
			case <-d.isUpdated:
				// debug
				log.Println("call dump")
				d.dump()
			case <-d.stop:
				log.Println("stop")
				return
			default:
				// debug
				log.Println("call dump")
				log.Println("DUMPER: no updates for dump")
			}
		}
	}()
}

func (d *Dumper) ScheduleUpdate() {
	if len(d.isUpdated) == 0 {
		// debug
		log.Println("flag up")
		d.isUpdated <- struct{}{}
	}
}

func (d *Dumper) dump() {
	if err := d.dumpFn(); err != nil {
		log.Println("DUMPER: unexpected error: ", err.Error())
	}
	log.Println("dump end")
}

func (d *Dumper) ScheduleStop() {
	d.stop <- struct{}{}
}
