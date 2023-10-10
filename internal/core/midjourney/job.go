package midjourney

import "github.com/Vaayne/aienvoy/internal/pkg/dtoutils"

const (
	MJGenerate   string = "Generate"
	MJUpscale    string = "Upscale"
	MJMaxUpscale string = "max_upscale"
	MJVariate    string = "Variate"
	MJReset      string = "Reset"
)

// JobMessage is a struct that represents the mid-journey job message.
// It contains the following fields:
// - Id: ID of the mid-journey job for the database
// - Action: action to be taken for the mid-journey job
// - Prompt: the text prompt for the mid-journey job
// - MessageImageIdx: index of the message image for the mid-journey job
// - MessageId: ID of the message for the mid-journey job
// - MessageHash: the hash code of the message for the mid-journey job
// - Channel: channel for the mid-journey job, it should match the message's channel
type JobMessage struct {
	Id              string
	Action          string
	Prompt          string
	MessageImageIdx int64
	MessageId       string
	MessageHash     string
	Channel         int64
	SourceClient    string
}

func (m JobMessage) toDTO() MjDTO {
	return MjDTO{
		BaseModel: dtoutils.BaseModel{
			Id: m.Id,
		},
		Prompt:          &m.Prompt,
		Action:          &m.Action,
		MessageImageIdx: &m.MessageImageIdx,
		MessageID:       &m.MessageId,
		MessageHash:     &m.MessageHash,
		ChannelID:       &m.Channel,
	}
}
