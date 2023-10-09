package midjourney

import (
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/internal/pkg/config"

	"github.com/bwmarrin/discordgo"
)

type Scene string

const (
	GenerateStart      Scene = "GenerateStart"
	GenerateProcessing Scene = "GenerateProcessing"
	GenerateEnd        Scene = "GenerateEnd"
	GenerateEditError  Scene = "GenerateEditError"
)

func CreateMsg(s *discordgo.Session, m *discordgo.MessageCreate) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic", "err", err, "channel_id", m.ChannelID, "content", m.Content, "type", m.Type)
		}
	}()

	// filter channel message
	if m.ChannelID != config.GetConfig().MidJourney.DiscordChannelId {
		return
	}

	// filter user message
	if m.Author == nil || m.Author.ID == s.State.User.ID {
		return
	}

	slog.Debug("discord new message", "channel_id", m.ChannelID, "content", m.Content, "type", m.Type)

	// when midjourney bot start to process, the message will contain "(Waiting to start)"
	if strings.Contains(m.Content, "(Waiting to start)") && !strings.Contains(m.Content, "Rerolling **") {
		// Notify(m.Message, GenerateStart, uuid.Nil, "")
		slog.Info("GenerateStart", "message", m.Content)
		return
	}

	// when midjourney bot end to process, the message will contain attachments
	for _, attachment := range m.Attachments {
		if attachment.Width > 0 && attachment.Height > 0 {
			slog.Info("GenerateEnd", "message", m.Content)
			return
		}
	}
}

func UpdateMsg(s *discordgo.Session, m *discordgo.MessageUpdate) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic", "err", err, "channel_id", m.ChannelID, "content", m.Content, "type", m.Type)
		}
	}()

	// filter channel message
	if m.ChannelID != config.GetConfig().MidJourney.DiscordChannelId {
		return
	}

	// filter user message
	if m.Author == nil || m.Author.ID == s.State.User.ID {
		return
	}

	slog.Debug("discord update message", "channel_id", m.ChannelID, "content", m.Content, "type", m.Type)

	// if message contain "(Stopped)", it means midjourney bot failed
	if strings.Contains(m.Content, "(Stopped)") {
		slog.Info("GenerateEnd", "message", m.Content)
		return
	}

	// if message contain "(0%)" or "(31%)" or "(62%)", it means midjourney bot is still processing and return the percentage
	// re := regexp.MustCompile(`\([1-9][0-9]?\%\)`)
	// if re.MatchString(m.Content) {
	// 	Notify(m.Message, GenerateProcessing, uuid.Nil, "")
	// 	return
	// }
}
