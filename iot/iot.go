package iot

import "time"

const (
	DateLayout = "2.1.2006"
	TimeLayout = "15:04:05"
)

type (
	IOTStorage interface {
		SaveValues(t time.Time, ID string, value interface{}) error
		GetDayValues(t time.Time) (map[string][]StorageValue, error)
	}

	StorageValue struct {
		Time  time.Time
		Value interface{}
	}

	IOTStorageMap struct {
		Values map[string]map[string][]StorageValue
	}
)

func NewIOTStorageMap() IOTStorage {
	return &IOTStorageMap{
		make(map[string]map[string][]StorageValue),
	}
}

func (s *IOTStorageMap) SaveValues(t time.Time, ID string, value interface{}) error {
	date := t.Format(DateLayout)
	if s.Values[date] == nil {
		s.Values[date] = make(map[string][]StorageValue)
	}

	s.Values[date][ID] = append(s.Values[date][ID], StorageValue{
		Time:  t,
		Value: value,
	})

	return nil
}

func (s *IOTStorageMap) GetDayValues(t time.Time) (map[string][]StorageValue, error) {
	dayValues := s.Values[t.Format(DateLayout)]
	return dayValues, nil
}
