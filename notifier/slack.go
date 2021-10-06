package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/logger"
	"github.com/syaiful6/payung/version"
)

type Slack struct {
	Base
	webhookURL string
	channel    string
	iconEmoji  string
	httpClient *http.Client
}

type SlackAttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type SlackAttachment struct {
	Fallback string                 `json:"fallback"`
	Text     string                 `json:"text"`
	Color    string                 `json:"color"`
	Fields   []SlackAttachmentField `json:"fields,omitempty"`
}

type SlackNotification struct {
	Text        string            `json:"text"`
	Channel     string            `json:"channel"`
	IconEmoji   string            `json:"icon_emoji"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

func (ctx *Slack) open() error {
	ctx.viper.SetDefault("icon_emoji", "floppy_disk")
	ctx.webhookURL = ctx.viper.GetString("webhook_url")
	ctx.channel = ctx.viper.GetString("channel")
	ctx.iconEmoji = ctx.viper.GetString("icon_emoji")
	ctx.httpClient = &http.Client{Transport: http.DefaultTransport}

	return nil
}

func (ctx *Slack) close() error {
	// Close idle connection
	ctx.httpClient.CloseIdleConnections()
	return nil
}

func (ctx *Slack) notify(info config.ModelRunInfo) error {
	logger.Info("-> Slack notifying...")
	status := ctx.getExitStatusName(info.ExitStatus)
	statusData := ctx.statusDataFor(info.ExitStatus)
	if statusData == nil {
		return fmt.Errorf("invalid status %s", status)
	}
	data := &SlackNotification{
		Text:        fmt.Sprintf("%s %s", statusData.Message, ctx.model.Name),
		Channel:     ctx.channel,
		IconEmoji:   ctx.iconEmoji,
		Attachments: ctx.getAttachments(info),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", ctx.webhookURL, bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ctx.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		logger.Info("Slack notifier: return status code %d", resp.StatusCode)
	}
	logger.Info("-> Slack notification completed")
	return nil
}

func (ctx *Slack) getAttachments(info config.ModelRunInfo) []SlackAttachment {
	elapsed := info.FinishedAt.Sub(info.StartedAt)
	return []SlackAttachment{
		{
			Fallback: fmt.Sprintf("%s - Job %s", ctx.statusText(info.ExitStatus), ctx.model.Name),
			Text:     ctx.statusText(info.ExitStatus),
			Color:    ctx.colorStatus(info.ExitStatus),
			Fields: []SlackAttachmentField{
				{
					Title: "Job",
					Value: ctx.model.Name,
					Short: false,
				},
				{
					Title: "Started",
					Value: info.StartedAt.Format("Mon, 2 Jan 2006 15:04:05 MST"),
					Short: false,
				},
				{
					Title: "Finished",
					Value: info.FinishedAt.Format("Mon, 2 Jan 2006 15:04:05 MST"),
					Short: false,
				},
				{
					Title: "Duration",
					Value: elapsed.Round(time.Minute).String(),
					Short: false,
				},
				{
					Title: "Version",
					Value: version.Version + "+" + version.Revision,
					Short: false,
				},
			},
		},
	}
}

func (ctx Slack) colorStatus(exitStatus int) string {
	switch exitStatus {
	case 0:
		return "good"
	case 1:
		return "danger"
	case 2:
		return "warning"
	default:
		return "danger"
	}
}

func (ctx *Slack) statusText(exitStatus int) string {
	switch exitStatus {
	case 0:
		return "Backup completed successfully!"
	case 1:
		return "Backup failed!"
	case 2:
		return "Backup interupted!"
	default:
		return fmt.Sprintf("Backup unknown exit status: %d", exitStatus)
	}
}
