package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gaoyulin/prometheus-webhook-dingtalk/models"
	"github.com/gaoyulin/prometheus-webhook-dingtalk/nacos"
	"github.com/gaoyulin/prometheus-webhook-dingtalk/template"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

func BuildDingTalkNotification(promMessage *models.WebhookMessage) (*models.DingTalkNotification, error) {

	title, err := template.ExecuteTextString(`{{ template "ding.link.title" . }}`, promMessage)
	if err != nil {
		return nil, err
	}
	content, err := template.ExecuteTextString(`{{ template "ding.link.content" . }}`, promMessage)
	if err != nil {
		return nil, err
	}
	var buttons []models.DingTalkNotificationButton
	var applicationName string
	for i, alert := range promMessage.Alerts.Firing() {
		buttons = append(buttons, models.DingTalkNotificationButton{
			Title:     fmt.Sprintf("Graph for alert #%d", i+1),
			ActionURL: alert.GeneratorURL,
		})
		applicationNames := alert.Labels.Values()
		for j, name := range applicationNames {
			print(j)
			if v, ok := nacos.GetMobiles(applicationName); ok {
				content = content + v
			}
			content = content
			applicationName = name
		}
	}

	notification := &models.DingTalkNotification{
		MessageType: "markdown",
		Markdown: &models.DingTalkNotificationMarkdown{
			Title: title,
			Text:  content,
		},
	}

	notification.At = new(models.DingTalkNotificationAt)
	if v, ok := nacos.GetMobiles(applicationName); ok {
		notification.At.AtMobiles = strings.Split(strings.TrimSpace(strings.Replace(strings.TrimSpace(v), "@", "", -1)), ",")
	}
	if _, ok := map[string]string(promMessage.CommonLabels)["is_at_all"]; ok {
		notification.At.IsAtAll = true
	}
	return notification, nil
}

func SendDingTalkNotification(httpClient *http.Client, webhookURL string, notification *models.DingTalkNotification) (*models.DingTalkNotificationResponse, error) {
	body, err := json.Marshal(&notification)
	if err != nil {
		return nil, errors.Wrap(err, "error encoding DingTalk request")
	}
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

	var robotResp models.DingTalkNotificationResponse
	enc := json.NewDecoder(req.Body)
	if err := enc.Decode(&robotResp); err != nil {
		return nil, errors.Wrap(err, "error decoding response from DingTalk")
	}
	return &robotResp, nil
}
