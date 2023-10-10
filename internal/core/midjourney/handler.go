package midjourney

import (
	"fmt"
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
	if m.ChannelID != fmt.Sprintf("%d", config.GetConfig().MidJourney.DiscordChannelId) {
		return
	}

	// filter user message
	if m.Author == nil || m.Author.ID == s.State.User.ID {
		return
	}

	slog.Debug("discord new message", "channel_id", m.ChannelID, "content", m.Content, "type", m.Type)

	dto, err := GetProcessingJobRecordByChannelIDAndStatus(defaultClient.Dao, m.ChannelID, StatusPending)
	if err != nil {
		slog.Error("get processing job by channel id error", "channel_id", m.ChannelID, "err", err)
		return
	}
	dto.MessageContent = &m.Content
	// when midjourney bot start to process, the message will contain "(Waiting to start)"
	if strings.Contains(m.Content, "(Waiting to start)") && !strings.Contains(m.Content, "Rerolling **") {
		// Notify(m.Message, GenerateStart, uuid.Nil, "")
		slog.Info("GenerateStart", "message", m.Content)
		status := StatusProcessing
		dto.Status = &status
		if _, err := UpdateJobRecord(defaultClient.Dao, *dto); err != nil {
			slog.Error("generate start update job record error", "err", err)
		}
		return
	}

	// when midjourney bot end to process, the message will contain attachments
	for _, attachment := range m.Attachments {
		if attachment.Width > 0 && attachment.Height > 0 {
			slog.Info("GenerateEnd", "message", m.Content)
			status := StatusCompleted
			dto.Status = &status

			// update image meta
			attachment := m.Attachments[0]
			dto.ImageName = &attachment.Filename
			dto.ImageUrl = &attachment.URL
			dto.ImageContentType = &attachment.ContentType
			size := int64(attachment.Size)
			dto.ImageSize = &size
			height := int64(attachment.Height)
			dto.ImageHeight = &height
			width := int64(attachment.Width)
			dto.ImageWidth = &width

			// update dto with message meta
			dto.MessageID = &m.ID
			hash := generateDiscordMsgHash(attachment.URL)
			dto.MessageHash = &hash
			if _, err := UpdateJobRecord(defaultClient.Dao, *dto); err != nil {
				slog.Error("generate end update job record error", "err", err)
			}
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
	if m.ChannelID != fmt.Sprintf("%d", config.GetConfig().MidJourney.DiscordChannelId) {
		return
	}

	// filter user message
	if m.Author == nil || m.Author.ID == s.State.User.ID {
		return
	}

	slog.Debug("discord update message", "channel_id", m.ChannelID, "content", m.Content, "type", m.Type)

	dto, err := GetProcessingJobRecordByChannelIDAndStatus(defaultClient.Dao, m.ChannelID, StatusProcessing)
	if err != nil || dto == nil {
		slog.Error("get processing job by channel id error", "channel_id", m.ChannelID, "err", err, "dto", dto)
		return
	}
	dto.MessageContent = &m.Content

	// if message contain "(Stopped)", it means midjourney bot failed
	if strings.Contains(m.Content, "(Stopped)") {
		slog.Info("GenerateEnd error", "message", m.Content)
		status := StatusFailed
		dto.Status = &status
		if _, err := UpdateJobRecord(defaultClient.Dao, *dto); err != nil {
			slog.Error("generate faled update job record error", "err", err)
		}
		return
	}
	if _, err := UpdateJobRecord(defaultClient.Dao, *dto); err != nil {
		slog.Error("generate faled update job record error", "err", err)
	}
}

func generateDiscordMsgHash(url string) string {
	_parts := strings.Split(url, "_")
	return strings.Split(_parts[len(_parts)-1], ".")[0]
}
