package utils

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
)

func SendServerMessage(key string, title string, message string) {
	c, err := client.NewClient()
	if err != nil {
		return
	}
	var postArgs protocol.Args
	postArgs.Set("title", title) // Set post args
	postArgs.Set("desp", message)
	url := fmt.Sprintf("https://sctapi.ftqq.com/%s.send", key)
	status, _, _ := c.Post(context.Background(), nil, url, &postArgs)
	if status != 200 {

	}
}
