package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Vaayne/aienvoy/internal/pkg/config"

	"github.com/Vaayne/aienvoy/internal/core/midjourney"
	tb "gopkg.in/telebot.v3"
)

func OnMidJourneyImagine(c tb.Context) error {
	text := strings.TrimSpace(c.Data())
	if text == "" {
		return c.Send("empty prompt")
	}
	ctx := c.Get(config.ContextKeyContext).(context.Context)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	mj := midjourney.GetDefaultClient()
	job, err := mj.ProcessMessage(&midjourney.JobMessage{
		Action:  midjourney.MJGenerate,
		Prompt:  text,
		Channel: mj.Client.ChannelID,
	})
	if err != nil {
		return c.Send(fmt.Sprintf("create midjourney job error: %s", err))
	}

	msg, err := c.Bot().Send(c.Sender(), "MidJourney job started, please wait ...")
	if err != nil {
		return fmt.Errorf("send message to user err: %v", err)
	}
	for {
		select {
		case <-ctx.Done():
			return c.Reply("time out error")
		default:
			time.Sleep(10 * time.Second)
			job, err = midjourney.GetJobRecord(mj.Dao, job.Id)
			if err != nil {
				slog.Error("could not get midjourney job record by id ")
				_, err = c.Bot().Edit(msg, fmt.Sprintf("could not get midjourney job record by id %s", job.Id))
				slog.Error("edit msg error", "err", err)
				return err
			}
			switch *job.Status {
			case midjourney.StatusProcessing:
				if job.MessageContent != nil {
					msgText := *job.MessageContent
					if _, err = c.Bot().Edit(msg, msgText); err != nil {
						slog.Warn("edit message error", "err", err, "text", msgText)
					}
				}
			case midjourney.StatusCompleted:
				slog.Info("midjourney generate completion")
				sendPhoto := func() error {
					photo := new(tb.Photo)
					if job.TelegramFileId != nil {
						photo.File = tb.File{FileID: *job.TelegramFileId}
					} else {
						resp, err := http.Get(*job.ImageUrl)
						if err != nil {
							return err
						}
						defer resp.Body.Close()
						photo = &tb.Photo{File: tb.File{
							FileReader: resp.Body,
						}}
					}
					if err := c.Bot().Delete(msg); err != nil {
						return err
					}
					err = c.Reply(photo)
					if job.TelegramFileId == nil {
						job.TelegramFileId = &photo.FileID
						if _, err := midjourney.UpdateJobRecord(mj.Dao, *job); err != nil {
							slog.Error("update midjourney job record error", "err", err)
						}
					}
					return err
				}
				return sendPhoto()
			case midjourney.StatusFailed:
				_, err = c.Bot().Edit(msg, "MidJourney job error")
				return err
			}
		}
	}
}
