package dingrobot

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type Roboter interface {
	Send(msg interface{}) error
	SendMarkdown(title, text string) error
	SendText(content string, atMobiles []string, isAtAll bool) error
	SetSecret(secret string)
}

type Robot struct {
	webHook string
	secret  string
}

func NewRobot(webHook string) Roboter {
	return &Robot{webHook: webHook}
}

func (r *Robot) SetSecret(secret string) {
	r.secret = secret
}

func (r *Robot) SendMarkdown(title, text string) error {
	return r.Send(&markdownMessage{
		MsgType: msgTypeMarkdown,
		Markdown: markdownParams{
			Title: title,
			Text:  text,
		},
	})
}

func (r *Robot) SendText(content string, atMobiles []string, isAtAll bool) error {
	return r.Send(&textMessage{
		MsgType: msgTypeText,
		Text: textParams{
			Content: content,
		},
		At: atParams{
			AtMobiles: atMobiles,
			IsAtAll:   isAtAll,
		},
	})
}

type dingResponse struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

func (r *Robot) Send(msg interface{}) error {
	m, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	webURL := r.webHook
	if r.secret != "" {
		webURL += genSignedURL(r.secret)
	}
	resp, err := http.Post(webURL, "application/json", bytes.NewReader(m))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var dr dingResponse
	err = json.Unmarshal(data, &dr)
	if err != nil {
		return err
	}
	if dr.Errcode != 0 {
		return fmt.Errorf("dingrobot send failed: %v", dr.Errmsg)
	}

	return nil
}

func genSignedURL(secret string) string {
	timeStr := fmt.Sprintf("%d", time.Now().UnixNano()/1e6)
	sign := fmt.Sprintf("%s\n%s", timeStr, secret)
	signData := computeHmacSha256(sign, secret)
	encodeURL := url.QueryEscape(signData)
	return fmt.Sprintf("&timestamp=%s&sign=%s", timeStr, encodeURL)
}

func computeHmacSha256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
