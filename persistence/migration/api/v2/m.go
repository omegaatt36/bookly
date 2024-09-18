package v2

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type accountWithNullableUserID struct {
	ID     string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID string `gorm:"type:uuid"`
}

func (a *accountWithNullableUserID) TableName() string {
	return "accounts"
}

type account struct {
	ID     string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID string `gorm:"type:uuid,not null"`
}

// User represents the database model for a user
type User struct {
	ID        string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Disabled  bool   `gorm:"not null;default:false"`
	Name      string `gorm:"type:varchar(255);not null"`
	Nickname  string `gorm:"type:varchar(255)"`

	Accounts []account
}

// CreateUser defines the migration, which creates the users.
var CreateUser = gormigrate.Migration{
	ID: "2024-09-18:create-users",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&User{}); err != nil {
			return err
		}

		if err := tx.Migrator().AddColumn(&accountWithNullableUserID{}, "user_id"); err != nil {
			return err
		}

		var accounts []accountWithNullableUserID
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Find(&accounts).Error; err != nil {
			return err
		}

		if len(accounts) != 0 {
			user := User{
				Name: "default",
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}

			if err := tx.Model(&account{}).Where("1 = 1").Update("user_id", user.ID).Error; err != nil {
				return err
			}
		}

		return tx.Migrator().AlterColumn(&account{}, "user_id")
	},
	Rollback: func(tx *gorm.DB) error {
		if err := tx.Migrator().DropColumn(&account{}, "user_id"); err != nil {
			return err
		}

		return tx.Migrator().DropTable(&User{})
	},
}
