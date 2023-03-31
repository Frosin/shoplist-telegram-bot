package metrics

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	tempPath = "/sys/devices/virtual/thermal/thermal_zone0/temp"
)

func GetPiTemp(path *string) (float64, error) {
	if path == nil {
		path = &tempPath
	}

	tempFile, err := os.Open(*path)
	if err != nil {
		return 0, fmt.Errorf("getPiTemp: open tempFile error: %w", err)
	}
	defer tempFile.Close()

	tempByte, err := io.ReadAll(tempFile)
	if err != nil {
		return 0, fmt.Errorf("getPiTemp: readAll error: %w", err)
	}
	tempStr := strings.TrimSpace(string(tempByte))

	if len(tempStr) != 5 {
		return 0, fmt.Errorf("getPiTemp: temperature invalid length: %s", tempStr)
	}

	base, err := strconv.ParseFloat(tempStr[:2], 64)
	if err != nil {
		return 0, fmt.Errorf("getPiTemp: parse base error: %w", err)
	}
	afterDot, err := strconv.ParseFloat(tempStr[2:3], 64)
	if err != nil {
		return 0, fmt.Errorf("getPiTemp: parse base error: %w", err)
	}
	return base + afterDot/10, nil
}
