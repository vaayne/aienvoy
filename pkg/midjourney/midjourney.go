package midjourney

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/bwmarrin/discordgo"
)

const (
	url       = "https://discord.com/api/v9/interactions"
	mjId      = "938956540159881230"
	mjAppID   = "936929561302675456"
	mjVersion = "1118961510123847772"
)

type MidJourney struct {
	mu            sync.Mutex
	GuildID       string        // DiscordServerId
	ChannelID     int64         // DiscordChannelId
	ApplicationId string        // DiscordAppId
	SessionId     string        // DiscordSessionId
	UserToken     string        // DiscordUserToken
	BotToken      string        // DiscordBotToken
	handlers      []interface{} // discord message handlers
}

func New(cfg config.MidJourney, handlers ...any) *MidJourney {
	mj := &MidJourney{
		mu:            sync.Mutex{},
		GuildID:       cfg.DiscordServerId,
		ChannelID:     cfg.DiscordChannelId,
		ApplicationId: cfg.DiscordAppId,
		SessionId:     cfg.DiscordSessionId,
		UserToken:     cfg.DiscordUserToken,
		BotToken:      cfg.DiscordBotToken,
	}
	mj.handlers = append(mj.handlers, handlers...)
	slog.Debug("create new midjourney client")
	return mj
}

func (m *MidJourney) Serve() {
	c, err := discordgo.New("Bot " + m.BotToken)
	if err != nil {
		panic(fmt.Sprintf("start new discord client error: %s", err))
	}
	if err := c.Open(); err != nil {
		panic(fmt.Sprintf("new discord client open connection error: %s", err))
	}
	for _, handler := range m.handlers {
		c.AddHandler(handler)
	}
	slog.Debug("start midjourney discord bot")
}

func (m *MidJourney) GenerateImage(prompt string) error {
	slog.Debug("MJ Generate", "prompt", prompt)
	requestBody := ReqTriggerDiscord{
		Type:          2,
		GuildId:       m.GuildID,
		ChannelId:     fmt.Sprintf("%d", m.ChannelID),
		ApplicationId: m.ApplicationId,
		SessionId:     m.SessionId,
		Data: DSCommand{
			Version: mjVersion,
			Id:      mjId,
			Name:    "imagine",
			Type:    1,
			Options: []DSOption{{Type: 3, Name: "prompt", Value: prompt}},
			ApplicationCommand: DSApplicationCommand{
				Id:                       mjId,
				ApplicationId:            mjAppID,
				Version:                  mjVersion,
				DefaultPermission:        true,
				DefaultMemberPermissions: nil,
				Type:                     1,
				Nsfw:                     false,
				Name:                     "imagine",
				Description:              "Lucky you!",
				DmPermission:             true,
				Options:                  []DSCommandOption{{Type: 3, Name: "prompt", Description: "The prompt to imagine", Required: true}},
			},
			Attachments: []interface{}{},
		},
	}
	return m.request(requestBody)
}

func (m *MidJourney) Upscale(index int64, messageId string, messageHash string) error {
	slog.Debug("MJ Upscale", "index", index, "messageId", messageId, "messageHash", messageHash)
	requestBody := ReqUpscaleDiscord{
		Type:          3,
		GuildId:       m.GuildID,
		ChannelId:     fmt.Sprintf("%d", m.ChannelID),
		ApplicationId: m.ApplicationId,
		SessionId:     m.SessionId,
		MessageFlags:  0,
		MessageId:     messageId,
		Data: UpscaleData{
			ComponentType: 2,
			CustomId:      fmt.Sprintf("MJ::JOB::upsample::%d::%s", index, messageHash),
		},
	}
	return m.request(requestBody)
}

func (m *MidJourney) MaxUpscale(messageId string, messageHash string) error {
	slog.Debug("MJ Max Upscale", "messageId", messageId, "messageHash", messageHash)
	requestBody := ReqUpscaleDiscord{
		Type:          3,
		GuildId:       m.GuildID,
		ChannelId:     fmt.Sprintf("%d", m.ChannelID),
		ApplicationId: m.ApplicationId,
		SessionId:     m.SessionId,
		MessageFlags:  0,
		MessageId:     messageId,
		Data: UpscaleData{
			ComponentType: 2,
			CustomId:      fmt.Sprintf("MJ::JOB::variation::1::%s::SOLO", messageHash),
		},
	}

	return m.request(requestBody)
}

func (m *MidJourney) Variate(index int64, messageId string, messageHash string) error {
	slog.Debug("MJ Variate", "index", index, "messageId", messageId, "messageHash", messageHash)
	requestBody := ReqVariationDiscord{
		Type:          3,
		GuildId:       m.GuildID,
		ChannelId:     fmt.Sprintf("%d", m.ChannelID),
		ApplicationId: m.ApplicationId,
		SessionId:     m.SessionId,
		MessageFlags:  0,
		MessageId:     messageId,
		Data: UpscaleData{
			ComponentType: 2,
			CustomId:      fmt.Sprintf("MJ::JOB::variation::%d::%s", index, messageHash),
		},
	}
	return m.request(requestBody)
}

func (m *MidJourney) Reset(messageId string, messageHash string) error {
	slog.Debug("MJ Reset", "messageId", messageId, "messageHash", messageHash)
	requestBody := ReqResetDiscord{
		Type:          3,
		GuildId:       m.GuildID,
		ChannelId:     fmt.Sprintf("%d", m.ChannelID),
		ApplicationId: m.ApplicationId,
		SessionId:     m.SessionId,
		MessageFlags:  0,
		MessageId:     messageId,
		Data: UpscaleData{
			ComponentType: 2,
			CustomId:      fmt.Sprintf("MJ::JOB::reroll::0::%s::SOLO", messageHash),
		},
	}
	return m.request(requestBody)
}

func (m *MidJourney) request(params any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	requestData, err := json.Marshal(params)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", m.UserToken)
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	slog.Info("MJ Response", "response", string(body), "error", err, "status", response.Status)
	return err
}
