package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func SendServerMessage(key string, title string, message string) {
	url := fmt.Sprintf("https://sctapi.ftqq.com/%s.send", key)
	data := map[string]string{"title": title, "desp": message}
	payload, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if resp.StatusCode != 200 || err != nil {
		fmt.Printf("发送消息失败,status code:%d,err:%s", resp.StatusCode, err.Error())
	}
	defer resp.Body.Close()
}
