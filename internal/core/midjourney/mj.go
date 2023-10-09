package midjourney

import (
	"sync"

	"github.com/pocketbase/pocketbase/daos"

	"github.com/Vaayne/aienvoy/internal/pkg/config"
	"github.com/Vaayne/aienvoy/pkg/midjourney"
)

var (
	once   sync.Once
	client *MidJourney
)

type MidJourney struct {
	Dao     *daos.Dao
	Service *midjourney.MidJourney
}

func New(dao *daos.Dao) *MidJourney {
	if client == nil {
		once.Do(func() {
			cfg := config.GetConfig().MidJourney
			mj := midjourney.New(cfg, CreateMsg, UpdateMsg)
			client = &MidJourney{
				Service: mj,
				Dao:     dao,
			}
		})
	}
	return client
}
