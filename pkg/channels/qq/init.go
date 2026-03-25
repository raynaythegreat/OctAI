package qq

import (
	"github.com/raynaythegreat/ai-business-hq/pkg/bus"
	"github.com/raynaythegreat/ai-business-hq/pkg/channels"
	"github.com/raynaythegreat/ai-business-hq/pkg/config"
)

func init() {
	channels.RegisterFactory("qq", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewQQChannel(cfg.Channels.QQ, b)
	})
}
