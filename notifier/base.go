package notifier

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/syaiful6/payung/config"
	"github.com/syaiful6/payung/logger"
)

// Base notifier
type Base struct {
	model        config.ModelConfig
	notifyConfig config.SubConfig
	viper        *viper.Viper
	name         string
}

type StatusData struct {
	Message string `json:"message"`
	Key     string `json:"key"`
}

func (base Base) getExitStatusName(exitStatus int) string {
	switch exitStatus {
	case 0:
		return "success"
	case 1:
		return "failed"
	case 2:
		return "interupted"
	default:
		return ""
	}
}

func (base Base) statusDataFor(exitStatus int) *StatusData {
	status := base.getExitStatusName(exitStatus)
	switch exitStatus {
	case 0:
		return &StatusData{
			Message: "Backup::Success",
			Key:     status,
		}
	case 1:
		return &StatusData{
			Message: "Backup::Failed",
			Key:     status,
		}
	case 2:
		return &StatusData{
			Message: "Backup::Interupted",
			Key:     status,
		}
	default:
		return nil
	}
}

func inStatus(statusCheck string, statuses []string) bool {
	for _, status := range statuses {
		if status == statusCheck {
			return true
		}
	}

	return false
}

type Context interface {
	open() error
	close() error
	notify(info config.ModelRunInfo) error
}

func newBase(model config.ModelConfig, notifyConfig config.SubConfig) Base {
	return Base{
		model:        model,
		notifyConfig: notifyConfig,
		viper:        notifyConfig.Viper,
		name:         notifyConfig.Name,
	}
}

func runNotifier(model config.ModelConfig, notifyConfig config.SubConfig, info config.ModelRunInfo) (err error) {
	base := newBase(model, notifyConfig)
	status := base.getExitStatusName(info.ExitStatus)
	var ctx Context
	switch notifyConfig.Type {
	case "slack":
		ctx = &Slack{Base: base}
	default:
		return fmt.Errorf("[%s] storage type has not implement", notifyConfig.Type)
	}

	sendOn := notifyConfig.Viper.GetStringSlice("send_on")
	if !inStatus(status, sendOn) {
		logger.Info("%s not send because %s not in send_on config", notifyConfig.Type, status)
		return nil
	}

	if err = ctx.open(); err != nil {
		return err
	}
	defer ctx.close()

	if err = ctx.notify(info); err != nil {
		return err
	}

	return nil
}

func Notify(model config.ModelConfig, info config.ModelRunInfo) error {
	if len(model.Notifiers) == 0 {
		return nil
	}

	logger.Info("Notify")
	for _, notifyConfig := range model.Notifiers {
		if err := runNotifier(model, notifyConfig, info); err != nil {
			return err
		}
	}
	logger.Info("Notify completed")

	return nil
}
