package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type Update struct {
	UpdateId int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Chat struct {
	Id int `json:"id"`
}

type Message struct {
	Chat   Chat   `json:"chat"`
	ChatId int    `json:"chat_id"`
	Text   string `json:"text"`
}

type Response[T any] struct {
	Ok     bool `json:"ok"`
	Result T    `json:"result"`
}

type Bot struct {
	Token          string
	LastUpdateId   int
	AllowedUpdates []string
}

type UpdatesRequest struct {
	Offset         int      `json:"offset"`
	Timeout        int      `json:"timeout"`
	AllowedUpdates []string `json:"allowed_updates"`
}

func (t *Bot) makeReq(method string, endpoint string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, t.getUrl(endpoint), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json")
	client := http.Client{}
	return client.Do(req)
}

func (t *Bot) getUrl(endpoint string) string {
	return "https://api.telegram.org/bot" + t.Token + endpoint
}

func (t *Bot) SendMessage(chat int, text string) error {
	messageBytes, err := json.Marshal(Message{ChatId: chat, Text: text})
	if err != nil {
		return err
	}
	res, err := t.makeReq("POST", "/sendMessage", messageBytes)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return errors.New("send message request status not 200")
	}
	return nil
}

func (t *Bot) GetUpdates() (Response[[]Update], error) {
	var output Response[[]Update]
	bodyBytes, err := json.Marshal(UpdatesRequest{Offset: t.LastUpdateId, AllowedUpdates: t.AllowedUpdates, Timeout: 60})
	if err != nil {
		return output, err
	}
	res, err := t.makeReq("POST", "/getUpdates", bodyBytes)
	if err != nil {
		return output, err
	}
	if res.StatusCode != 200 {
		return output, errors.New("get updates request status not 200")
	}
	err = json.NewDecoder(res.Body).Decode(&output)
	return output, err
}
