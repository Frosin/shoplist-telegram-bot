package iot

import (
	"log"
	"time"
)

const (
	DateLayout = "2.1.2006"
	TimeLayout = "15:04:05"
)

type (
	IOTStorage interface {
		SaveValues(t time.Time, ID string, value float64) error
		GetDayValues(t time.Time) (map[string][]StorageValue, error)
		GetCurrentValue(ID string) float64
	}

	StorageValue struct {
		Time  time.Time
		Value float64
	}

	IOTStorageMap struct {
		Values map[string]map[string][]StorageValue
		Value  map[string]float64
	}
)

func NewIOTStorageMap() IOTStorage {
	return &IOTStorageMap{
		Values: make(map[string]map[string][]StorageValue),
		Value:  make(map[string]float64),
	}
}

func (s *IOTStorageMap) SaveValues(t time.Time, ID string, value float64) error {
	date := t.Format(DateLayout)
	if s.Values[date] == nil {
		s.Values[date] = make(map[string][]StorageValue)
	}

	s.Values[date][ID] = append(s.Values[date][ID], StorageValue{
		Time:  t,
		Value: value,
	})

	// debug
	log.Println("SetCurrentValue", ID, value)
	//
	s.Value[ID] = value

	return nil
}

func (s *IOTStorageMap) GetDayValues(t time.Time) (map[string][]StorageValue, error) {
	dayValues := s.Values[t.Format(DateLayout)]
	return dayValues, nil
}

func (s *IOTStorageMap) GetCurrentValue(ID string) float64 {
	// debug
	log.Println("GetCurrentValue", ID, s.Value[ID])
	//
	return s.Value[ID]
}
