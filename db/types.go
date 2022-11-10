package db

import "time"

type App struct {
	ID              string    `db:"id"`
	SystemID        string    `db:"system_id"`
	Name            string    `db:"name"`
	Description     string    `db:"description"`
	WikiURL         string    `db:"wiki_url"`
	IntegrationDate time.Time `db:"integration_date"`
	EditedDate      time.Time `db:"edited_date"`
	Username        string    `db:"username"`
	JobCount        int       `db:"job_count"`
	IsFavorite      bool      `db:"is_favorite"`
	IsPublic        bool      `db:"is_public"`
}
