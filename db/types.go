package db

import (
	"database/sql"
)

type App struct {
	ID              string         `db:"id"`
	SystemID        string         `db:"system_id"`
	Name            string         `db:"name"`
	Description     sql.NullString `db:"description"`
	WikiURL         sql.NullString `db:"wiki_url"`
	IntegrationDate sql.NullTime   `db:"integration_date"`
	EditedDate      sql.NullTime   `db:"edited_date"`
	Username        sql.NullString `db:"username"`
	JobCount        int            `db:"job_count"`
	IsFavorite      bool           `db:"is_favorite"`
	IsPublic        bool           `db:"is_public"`
}
