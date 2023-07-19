package dao

import (
	"aienvoy/internal/pkg/context"

	"github.com/pocketbase/pocketbase/daos"
)

type Dao struct {
	ctx context.Context
}

func New(ctx context.Context) *Dao {
	return &Dao{ctx: ctx}
}

func (d *Dao) RunInTransaction(fn func(tx *daos.Dao) error) error {
	return d.ctx.Dao().RunInTransaction(func(tx *daos.Dao) error {
		return fn(tx)
	})
}
