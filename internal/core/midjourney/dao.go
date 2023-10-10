package midjourney

import (
	"github.com/Vaayne/aienvoy/internal/pkg/dtoutils"
	"github.com/google/uuid"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/tools/types"
	"log/slog"
)

const tableName = "midjourney_jobs"

const (
	StatusPending    = "Pending"
	StatusProcessing = "Pending"
	StatusCompleted  = "Completed"
	StatusFailed     = "Failed"
)

type MjDTO struct {
	dtoutils.BaseModel
	Prompt           *string `json:"prompt,omitempty" mapstructure:"prompt,omitempty" db:"prompt"`
	Action           *string `json:"action,omitempty" mapstructure:"action,omitempty" db:"action"`
	Status           *string `json:"status,omitempty" mapstructure:"status,omitempty" db:"status"`
	ChannelID        *int64  `json:"channel_id,omitempty" mapstructure:"channel_id,omitempty" db:"channel_id"`
	MessageImageIdx  *int64  `json:"message_image_idx,omitempty" mapstructure:"message_image_idx,omitempty" db:"message_image_idx"`
	MessageID        *string `json:"message_id,omitempty" mapstructure:"message_id,omitempty" db:"message_id"`
	MessageHash      *string `json:"message_hash,omitempty" mapstructure:"message_hash,omitempty" db:"message_hash"`
	MessageContent   *string `json:"message_content,omitempty" mapstructure:"message_content,omitempty" db:"message_content"`
	ImageName        *string `json:"image_name,omitempty" mapstructure:"image_name,omitempty" db:"image_name"`
	ImageUrl         *string `json:"image_url,omitempty" mapstructure:"image_url,omitempty" db:"image_url"`
	ImageContentType *string `json:"image_content_type,omitempty" mapstructure:"image_content_type,omitempty" db:"image_content_type"`
	ImageSize        *int64  `json:"image_size,omitempty" mapstructure:"image_size,omitempty" db:"image_size"`
	ImageHeight      *int64  `json:"image_height,omitempty" mapstructure:"image_height,omitempty" db:"image_height"`
	ImageWidth       *int64  `json:"image_width,omitempty" mapstructure:"image_width,omitempty" db:"image_width"`
}

func (m MjDTO) TableName() string {
	return tableName
}

//
//func (m MjDTO) LogValue() slog.Value {
//	return slog.GroupValue(
//		slog.String("id", m.Id),
//		slog.Time("created", m.Created.Time()),
//		slog.Time("updated", m.Updated.Time()),
//		slog.String("prompt", *m.Prompt),
//		slog.String("action", *m.Action),
//		slog.String("status", *m.Status),
//		slog.Int64("channel_id", *m.ChannelID),
//		slog.Int64("message_image_idx", *m.MessageImageIdx),
//		slog.String("message_id", *m.MessageID),
//		slog.String("message_hash", *m.MessageHash),
//		slog.String("image_name", *m.ImageName),
//		slog.String("image_url", *m.ImageUrl),
//		slog.String("image_content_type", *m.ImageContentType),
//		slog.Int64("image_size", *m.ImageSize),
//		slog.Int64("image_height", *m.ImageHeight),
//		slog.Int64("image_width", *m.ImageWidth),
//	)
//}

func CreateJobRecord(tx *daos.Dao, mj MjDTO) (*MjDTO, error) {
	status := StatusPending
	mj.Status = &status
	mj.Id = uuid.New().String()
	mj.Created = types.NowDateTime()
	mj.Updated = types.NowDateTime()
	if err := tx.DB().Model(&mj).Insert(); err != nil {
		return nil, err
	}
	return GetJobRecord(tx, mj.Id)
}

func UpdateJobRecord(tx *daos.Dao, mj MjDTO) (*MjDTO, error) {
	mj.Updated = types.NowDateTime()
	if err := tx.DB().Model(&mj).Update(); err != nil {
		return nil, err
	}
	return GetJobRecord(tx, mj.Id)
}

func GetJobRecord(tx *daos.Dao, id string) (*MjDTO, error) {
	var dto MjDTO
	if err := tx.DB().Select().Model(id, &dto); err != nil {
		return nil, err
	}
	slog.Debug("get job record", "dto", &dto)
	return &dto, nil
}

func GetProcessingJobRecordByChannelIDAndStatus(tx *daos.Dao, id, status string) (*MjDTO, error) {
	var dto MjDTO
	if err := tx.DB().
		Select().
		From(tableName).
		Where(dbx.HashExp{"channel_id": id, "status": status}).
		OrderBy("updated DESC").
		Limit(1).
		One(&dto); err != nil {
		return nil, err
	}

	slog.Debug("get job record by channel id and status", "dto", &dto)
	return &dto, nil
}
