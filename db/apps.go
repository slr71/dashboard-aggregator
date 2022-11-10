package db

import (
	"context"

	"github.com/doug-martin/goqu/v9"
)

type PopularFeaturedAppsConfig struct {
	Username          string
	GroupsIndex       int
	FeaturedAppIDs    []string
	StartDateInterval string
}

func (d *Database) PopularFeaturedApps(ctx context.Context, cfg *PopularFeaturedAppsConfig, opts ...QueryOption) ([]App, error) {
	var (
		err  error
		db   GoquDatabase
		apps []App
	)

	querySettings := &QuerySettings{}
	for _, opt := range opts {
		opt(querySettings)
	}

	if querySettings.tx != nil {
		db = querySettings.tx
	} else {
		db = d.goquDB
	}

	appListingT := goqu.T("app_listing")
	jobsT := goqu.T("jobs")
	usersT := goqu.T("users")
	workspaceT := goqu.T("workspace")
	appCatGroupT := goqu.T("app_category_group")
	appCatAppT := goqu.T("app_category_app")

	subquery := db.From(usersT).
		Join(workspaceT, goqu.On(usersT.Col("id").Eq(workspaceT.Col("user_id")))).
		Join(appCatGroupT, goqu.On(workspaceT.Col("root_category_id").Eq(appCatGroupT.Col("parent_category_id")))).
		Join(appCatAppT, goqu.On(appCatGroupT.Col("child_category_id").Eq(appCatAppT.Col("category_id")))).
		Where(
			usersT.Col("username").Eq(cfg.Username),
			appCatGroupT.Col("child_index").Eq(cfg.GroupsIndex),
			appCatAppT.Col("app_id").Eq(appListingT.Col("id")),
		)

	query := db.From(appListingT).
		Select(
			appListingT.Col("id"),
			goqu.L(`'de'`).As("system_id"),
			appListingT.Col("name"),
			appListingT.Col("description"),
			appListingT.Col("wiki_url"),
			appListingT.Col("integration_date"),
			appListingT.Col("edited_date"),
			appListingT.Col("integrator_username").As(goqu.C("username")),
			goqu.COUNT(jobsT.Col("id")).As(goqu.C("job_count").Eq(goqu.C)),
			goqu.L("EXISTS(?)", subquery).As("is_favorite"),
			goqu.L("true").As("is_public"),
		).
		Join(jobsT, goqu.On(jobsT.Col("app_id").Eq(goqu.Cast(appListingT.Col("id"), "TEXT")))).
		Where(
			appListingT.Col("id").Eq(goqu.Any(cfg.FeaturedAppIDs)),
			appListingT.Col("deleted").Eq(goqu.L("false")),
			appListingT.Col("disabled").Eq(goqu.L("false")),
			appListingT.Col("integration_date").IsNotNull(),
			goqu.Or(
				jobsT.Col("start_date").Gte(goqu.L("now() - ?", goqu.Cast(goqu.L(cfg.StartDateInterval), "interval"))),
				jobsT.Col("start_date").IsNull(),
			),
		).
		GroupBy(
			appListingT.Col("id"),
			appListingT.Col("name"),
			appListingT.Col("description"),
			appListingT.Col("wiki_url"),
			appListingT.Col("integration_date"),
			appListingT.Col("edited_date"),
			appListingT.Col("integrator_username"),
		).
		Order(
			goqu.C("job_count").Desc(),
		)

	if querySettings.hasLimit {
		query = query.Limit(querySettings.limit)
	}

	if querySettings.hasOffset {
		query = query.Offset(querySettings.offset)
	}

	executor := query.Executor()

	if err = executor.ScanStructsContext(ctx, &apps); err != nil {
		return nil, err
	}

	return apps, err
}
