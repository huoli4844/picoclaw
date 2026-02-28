package internalim

import (
	"context"
	"sync/atomic"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/media"
)

// ChannelInterface defines the interface for channels to avoid circular imports
type ChannelInterface interface {
	Name() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Send(ctx context.Context, msg bus.OutboundMessage) error
	IsRunning() bool
	IsAllowed(senderID string) bool
	IsAllowedSender(sender bus.SenderInfo) bool
	ReasoningChannelID() string
}

// PlaceholderRecorder defines the interface for placeholder recording
type PlaceholderRecorder interface {
	RecordPlaceholder(channel, chatID, placeholderID string)
	RecordTypingStop(channel, chatID string, stop func())
	RecordReactionUndo(channel, chatID string, undo func())
}

// BaseChannel is a simplified version of the channels.BaseChannel to avoid circular imports
type BaseChannel struct {
	Config              any
	Bus                 *bus.MessageBus
	running             atomic.Bool
	name                string
	AllowList           []string
	MaxMessageLength    int
	GroupTrigger        config.GroupTriggerConfig
	MediaStore          media.MediaStore
	PlaceholderRecorder PlaceholderRecorder
	Owner               ChannelInterface // the concrete channel that embeds this BaseChannel
	reasoningChannelID  string
}

// NewBaseChannel creates a new BaseChannel instance
func NewBaseChannel(
	name string,
	config any,
	bus *bus.MessageBus,
	allowList []string,
) *BaseChannel {
	return &BaseChannel{
		Config:    config,
		Bus:       bus,
		name:      name,
		AllowList: allowList,
	}
}

// Name returns the channel name
func (c *BaseChannel) Name() string {
	return c.name
}

// IsRunning returns whether the channel is running
func (c *BaseChannel) IsRunning() bool {
	return c.running.Load()
}

// SetRunning sets the running state
func (c *BaseChannel) SetRunning(running bool) {
	c.running.Store(running)
}

// IsAllowed checks if a sender ID is allowed
func (c *BaseChannel) IsAllowed(senderID string) bool {
	if len(c.AllowList) == 0 {
		return true
	}
	for _, allowed := range c.AllowList {
		if allowed == senderID {
			return true
		}
	}
	return false
}

// IsAllowedSender checks if a sender is allowed
func (c *BaseChannel) IsAllowedSender(sender bus.SenderInfo) bool {
	return c.IsAllowed(sender.PlatformID)
}

// ReasoningChannelID returns the reasoning channel ID
func (c *BaseChannel) ReasoningChannelID() string {
	return c.reasoningChannelID
}

// HandleMessage handles inbound messages
func (c *BaseChannel) HandleMessage(
	ctx context.Context,
	peer bus.Peer,
	messageID, senderID, chatID, content string,
	media []string,
	metadata map[string]string,
	senderOpts ...bus.SenderInfo,
) {
	// Use SenderInfo-based allow check when available, else fall back to string
	var sender bus.SenderInfo
	if len(senderOpts) > 0 {
		sender = senderOpts[0]
	}
	if sender.CanonicalID != "" || sender.PlatformID != "" {
		if !c.IsAllowedSender(sender) {
			return
		}
	} else {
		if !c.IsAllowed(senderID) {
			return
		}
	}

	// Create and publish inbound message
	inboundMsg := bus.InboundMessage{
		Peer:      peer,
		MessageID: messageID,
		Channel:   c.Name(),
		SenderID:  senderID,
		ChatID:    chatID,
		Content:   content,
		Media:     media,
		Metadata:  metadata,
		Sender:    sender,
	}

	c.Bus.PublishInbound(ctx, inboundMsg)
}
