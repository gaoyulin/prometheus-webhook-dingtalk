package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gaoyulin/prometheus-webhook-dingtalk/models"
	"github.com/gaoyulin/prometheus-webhook-dingtalk/template"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

func BuildDingTalkNotification(promMessage *models.WebhookMessage) (*models.DingTalkNotification, error) {
	println("===============BuildDingTalkNotification start===================================>")

	title, err := template.ExecuteTextString(`{{ template "ding.link.title" . }}`, promMessage)
	if err != nil {
		return nil, err
	}
	content, err := template.ExecuteTextString(`{{ template "ding.link.content" . }}`, promMessage)
	if err != nil {
		return nil, err
	}
	var buttons []models.DingTalkNotificationButton
	for i, alert := range promMessage.Alerts.Firing() {
		buttons = append(buttons, models.DingTalkNotificationButton{
			Title:     fmt.Sprintf("Graph for alert #%d", i+1),
			ActionURL: alert.GeneratorURL,
		})
	}
	notification := &models.DingTalkNotification{
		MessageType: "markdown",
		Markdown: &models.DingTalkNotificationMarkdown{
			Title: title,
			Text:  "@17313006617" + content,
		},
	}

	println(content)
	println("===============BuildDingTalkNotification 11111111===================================>")
	println(promMessage.CommonLabels.Values())
	println(promMessage.CommonLabels.Names())

	notification.At = new(models.DingTalkNotificationAt)
	if v, ok := map[string]string(promMessage.CommonLabels)["at_mobiles"]; ok {
		notification.At.AtMobiles = strings.Split(strings.TrimSpace(v), ",")
	}

	if _, ok := map[string]string(promMessage.CommonLabels)["is_at_all"]; ok {
		notification.At.IsAtAll = true
	}
	println("===============BuildDingTalkNotification 222222===================================>")

	return notification, nil
}

func SendDingTalkNotification(httpClient *http.Client, webhookURL string, notification *models.DingTalkNotification) (*models.DingTalkNotificationResponse, error) {
	body, err := json.Marshal(&notification)
	if err != nil {
		return nil, errors.Wrap(err, "error encoding DingTalk request")
	}

	print("==================================================>")

	httpReq, err := http.NewRequest("POST", webhookURL, bytes.NewReader(body))
	if err != nil {
		return nil, errors.Wrap(err, "error building DingTalk request")
	}
	httpReq.Header.Set("Content-Type", "application/json")

	req, err := httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.Wrap(err, "error sending notification to DingTalk")
	}
	defer req.Body.Close()

	if req.StatusCode != 200 {
		return nil, errors.Errorf("unacceptable response code %d", req.StatusCode)
	}

	print("==================================================>")

	var robotResp models.DingTalkNotificationResponse
	enc := json.NewDecoder(req.Body)
	if err := enc.Decode(&robotResp); err != nil {
		return nil, errors.Wrap(err, "error decoding response from DingTalk")
	}
	return &robotResp, nil
}
