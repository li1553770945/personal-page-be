package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func SendServerMessage(key string, title string, message string, logger *logrus.Logger) {
	url := fmt.Sprintf("https://sctapi.ftqq.com/%s.send", key)
	data := map[string]string{"title": title, "desp": message}
	payload, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if req == nil {
		logger.Error("发送消息失败，构建req为nil")
		return
	}
	if err != nil {
		logger.Error("发送消息失败，构建req失败", err.Error())
		return
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp == nil {
		logger.Error("发送消息失败，返回值为nil")
		return
	}
	if err != nil {
		logger.Error("发送消息失败:", err.Error())
		return
	}
	if resp.StatusCode != 200 {
		logger.Error("发送消息失败，状态码：", resp.StatusCode)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
}
