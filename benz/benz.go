package benz

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/benzstorage"
)

const (
	apiURL          = "https://gas-monitoring.admlr.lipetsk.ru/reports/data/"
	collectInterval = 30 * time.Minute
	httpTimeout     = 15 * time.Second
)

// GasStation mirrors the relevant fields from the gas monitoring API response.
type GasStation struct {
	ID        int    `json:"id"`
	AZS       string `json:"azs"`
	Address   string `json:"address"`
	IsWorking bool   `json:"is_working"`
	Has95     bool   `json:"has_95"`
	Queue     string `json:"queue"`
}

type apiResponse struct {
	Results []GasStation `json:"results"`
}

// FetchStations calls the API and returns the current station list.
func FetchStations() ([]GasStation, error) {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", apiURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var result apiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return result.Results, nil
}

// collect runs one fetch-and-store cycle.
func collect(storage *benzstorage.Storage) {
	stations, err := FetchStations()
	if err != nil {
		log.Printf("benz collector: fetch error: %v", err)
		return
	}

	inputs := make([]benzstorage.SnapshotInput, 0, len(stations))
	for _, st := range stations {
		inputs = append(inputs, benzstorage.SnapshotInput{
			AzsID:     st.ID,
			Address:   st.Address,
			AzsName:   st.AZS,
			IsWorking: st.IsWorking,
			Has95:     st.Has95,
			Queue:     st.Queue,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := storage.InsertSnapshots(ctx, inputs); err != nil {
		log.Printf("benz collector: insert error: %v", err)
		return
	}

	log.Printf("benz collector: saved %d snapshots", len(inputs))
}

// StartCollector starts the background goroutine that collects data every 30 min.
// It runs one collection immediately on startup.
func StartCollector(storage *benzstorage.Storage) {
	go func() {
		collect(storage)
		for range time.Tick(collectInterval) {
			collect(storage)
		}
	}()
}
