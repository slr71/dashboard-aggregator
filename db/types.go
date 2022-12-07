package db

import (
	"github.com/guregu/null"
)

type App struct {
	ID              string      `db:"id" json:"id"`
	SystemID        string      `db:"system_id" json:"system_id"`
	Name            string      `db:"name" json:"name"`
	Description     null.String `db:"description" json:"description"`
	WikiURL         null.String `db:"wiki_url" json:"wiki_url"`
	IntegrationDate null.Time   `db:"integration_date" json:"integration_date"`
	EditedDate      null.Time   `db:"edited_date" json:"edited_date"`
	Username        null.String `db:"username" json:"username"`
	JobCount        null.String `db:"job_count" json:"job_count"`
	IsFavorite      bool        `db:"is_favorite" json:"is_favorite"`
	IsPublic        bool        `db:"is_public" json:"is_public"`
}
