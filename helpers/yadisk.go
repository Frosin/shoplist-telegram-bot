package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

var (
	ynxUrl = "https://cloud-api.yandex.net/v1/disk"
)

type UploadError struct {
	Error       string `json:"error"`
	Description string `json:"description"`
}

type UploadInfo struct {
	URL            string `json:"href"`
	Method         string `json:"method"`
	URLIsTemplated bool   `json:"templated"`
}

func UploadBackupDB(token, uploadPath string, filePath string, overwriteFile bool) error {
	uploadRequestURL := "https://cloud-api.yandex.net/v1/disk/resources/upload"

	// Checking file existence before requesting
	normalizedFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("Failed to get path Abs: %w", err)
	}

	fileInfo, err := os.Stat(normalizedFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("File for uploading not found: %w", err)
		} else {
			return fmt.Errorf("Failed to stat uploading file: %w", err)
		}
	}

	if !fileInfo.Mode().IsRegular() {
		return errors.New("Only regular files uploading is supported right now")
	}

	client := http.Client{}
	uploadInfo := UploadInfo{}

	// The first request will get from Yandex upload URL
	req, err := http.NewRequest("GET", uploadRequestURL, nil)
	if err != nil {
		return fmt.Errorf("Failed to create get upload request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "OAuth "+token)

	query := url.Values{}
	query.Add("path", "disk:/"+uploadPath+"/"+fileInfo.Name())
	query.Add("overwrite", strconv.FormatBool(overwriteFile))

	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&uploadInfo)
		if err != nil {
			return fmt.Errorf("Failed to decode response: %w", err)
		}
	} else {
		errorData := UploadError{}
		err = json.NewDecoder(resp.Body).Decode(&errorData)
		if err != nil {
			return fmt.Errorf("Failed to decode response: %w", err)
		}
		return fmt.Errorf("Failed to get upload path: %s: %s", errorData.Error, errorData.Description)
	}

	if uploadInfo.URL == "" {
		return errors.New("Got empty upload URL")
	}

	file, err := os.Open(normalizedFilePath)
	if err != nil {
		return fmt.Errorf("Failed to open file: %w", err)
	}
	defer file.Close()

	uploadReq, err := http.NewRequest("PUT", uploadInfo.URL, file)
	if err != nil {
		return fmt.Errorf("Failed to create new PUT request: %w", err)
	}

	uploadResp, err := client.Do(uploadReq)
	if err != nil {
		return fmt.Errorf("Failed to send upload request: %w", err)
	}
	defer uploadResp.Body.Close()

	switch uploadResp.StatusCode {
	case http.StatusCreated:
		fallthrough
	case http.StatusAccepted:
		return nil
	case http.StatusRequestEntityTooLarge:
		return errors.New("File upload is too large.")
	case http.StatusInsufficientStorage:
		return errors.New("There is no space left on your Yandex.Disk.")
	default:
		return errors.New("Failed to upload file (error on Yandex's side). Try again later.")
	}
}
