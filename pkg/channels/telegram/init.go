package telegram

import (
	"github.com/raynaythegreat/ai-business-hq/pkg/bus"
	"github.com/raynaythegreat/ai-business-hq/pkg/channels"
	"github.com/raynaythegreat/ai-business-hq/pkg/config"
)

func init() {
	channels.RegisterFactory("telegram", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewTelegramChannel(cfg, b)
	})
}
