package midjourney

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/pocketbase/pocketbase/daos"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/midjourney"
)

var (
	once          sync.Once
	defaultClient *MidJourney
)

type MidJourney struct {
	Dao    *daos.Dao
	Client *midjourney.MidJourney
}

func New(dao *daos.Dao) *MidJourney {
	if defaultClient == nil {
		once.Do(func() {
			cfg := config.GetConfig().MidJourney
			mj := midjourney.New(cfg, CreateMsg, UpdateMsg)
			defaultClient = &MidJourney{
				Client: mj,
				Dao:    dao,
			}
		})
	}
	return defaultClient
}

func GetDefaultClient() *MidJourney {
	if defaultClient == nil {
		panic("Midjourney not init")
	}
	return defaultClient
}

func (m *MidJourney) ProcessMessage(msg *JobMessage) (job *MjDTO, err error) {
	if msg.Action != MJGenerate && m.Client.ChannelID != msg.Channel {
		slog.Error("channel id not match", "msg_channel_id", msg.Channel, "mj_channel_id", m.Client.ChannelID)
		return nil, errors.New("channel id not match")
	}

	ctx := context.Background()
	dto := msg.toDTO()
	if msg.Id != "" {
		job, err = UpdateJobRecord(m.Dao, dto)
		if err != nil {
			slog.Error("update job record error", "err", err, "msg", msg)
			return nil, err
		}
	} else {
		job, err = CreateJobRecord(m.Dao, dto)
		if err != nil {
			slog.Error("create job record error", "err", err, "msg", msg)
			return nil, err
		}
	}
	slog.InfoContext(ctx, "process message", "channel_id", msg.Channel, "content", msg.Prompt, "action", msg.Action, "message_id", msg.MessageId)
	switch msg.Action {
	case MJGenerate:
		err = m.Client.GenerateImage(msg.Prompt)
	case MJUpscale:
		err = m.Client.Upscale(msg.MessageImageIdx, msg.MessageId, msg.MessageHash)
	case MJVariate:
		err = m.Client.Variate(msg.MessageImageIdx, msg.MessageId, msg.MessageHash)
	case MJReset:
		err = m.Client.Reset(msg.MessageId, msg.MessageHash)
	}
	return
}
