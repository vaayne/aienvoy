package migrations

import (
	"github.com/Vaayne/aienvoy/internal/pkg/config"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		// add up queries...
		dao := daos.New(db)

		for _, admin := range config.GetConfig().Admins {
			user := &models.Admin{
				Email: admin.Email,
			}
			if err := user.SetPassword(admin.Password); err != nil {
				return err
			}
			if err := dao.SaveAdmin(user); err != nil {
				return err
			}
		}
		return nil
	}, func(db dbx.Builder) error {
		dao := daos.New(db)
		for _, admin := range config.GetConfig().Admins {
			user, _ := dao.FindAdminByEmail(admin.Email)
			if user != nil {
				return dao.DeleteAdmin(user)
			}
		}
		return nil
	})
}
