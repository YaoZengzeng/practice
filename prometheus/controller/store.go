package controller

import (
	"fmt"
)

type Store interface {
	GetPodMetricsRecords(id string) ([]string, error)

	AddPodMetricsRecords(id string, metrics []string) error

	DeletePodMetricsRecords(id string, metrics []string) error

	ResetPodMetricsRecords(id string, metrics []string) error
}

// memoryStore implements the Store interface, mainly used for testing.
type memoryStore struct {
	podMetricsRecords map[string]map[string]struct{}
}

func NewMemoryStore() Store {
	return &memoryStore{
		podMetricsRecords:	make(map[string]map[string]struct{}),
	}
}

func (m *memoryStore) GetPodMetricsRecords(id string) ([]string, error) {
	if m.podMetricsRecords[id] == nil {
		return nil, nil
	}

	var metrics []string
	for metric, _ := range m.podMetricsRecords[id] {
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (m *memoryStore) AddPodMetricsRecords(id string, metrics []string) error {
	if m.podMetricsRecords[id] == nil {
		m.podMetricsRecords[id] = make(map[string]struct{})
	}

	for _, metric := range metrics {
		m.podMetricsRecords[id][metric] = struct{}{}
	}

	fmt.Printf("%s\n%s\n", id, m.podMetricsRecords[id])

	return nil
}

func (m *memoryStore) DeletePodMetricsRecords(id string, metrics []string) error {
	return nil
}

func (m *memoryStore) ResetPodMetricsRecords(id string, metrics []string) error {
	return nil
}
