package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func SendServerMessage(key string, title string, message string, logger *logrus.Logger) {
	url := fmt.Sprintf("https://sctapi.ftqq.com/%s.send", key)
	data := map[string]string{"title": title, "desp": message}
	payload, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		logger.WithError(err).Error("send ServerChan message failed: build request")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.WithError(err).Error("send ServerChan message failed")
		return
	}
	if resp == nil {
		logger.Error("send ServerChan message failed: nil response")
		return
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		logger.WithError(readErr).Error("send ServerChan message failed: read response")
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(body),
		}).Error("send ServerChan message failed: bad status")
		return
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Error string `json:"error"`
			Errno int    `json:"errno"`
		} `json:"data"`
	}
	if len(body) > 0 && json.Unmarshal(body, &result) == nil {
		if result.Code != 0 || result.Data.Errno != 0 || (result.Data.Error != "" && result.Data.Error != "SUCCESS") {
			logger.WithFields(logrus.Fields{
				"code":    result.Code,
				"message": result.Message,
				"error":   result.Data.Error,
				"errno":   result.Data.Errno,
			}).Error("send ServerChan message failed: rejected")
		}
	}
}
