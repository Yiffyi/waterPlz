package waterplz

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type WeComBot struct {
	Key string
}

type weComResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (bot *WeComBot) SendText(msg string) error {
	return bot.SendMessage(map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": msg,
		},
	})
}

func (bot *WeComBot) SendMarkdown(md string) error {
	return bot.SendMessage(map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": md,
		},
	})
}

func (bot *WeComBot) SendMessage(msg map[string]interface{}) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("unable to marshal body: %w", err)
	}

	resp, err := http.Post("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key="+bot.Key, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("unable to perform send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bad HTTP Status: %s\n\t%s", resp.Status, string(b))
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read body: %w", err)
	}

	var r weComResponse
	err = json.Unmarshal(b, &r)
	if err != nil {
		return fmt.Errorf("unable to unmarshal body: %w", err)
	}

	if r.ErrCode != 0 {
		return errors.New(r.ErrMsg)
	}
	slog.Info("WeCom send", "body", string(b))
	return nil
}
