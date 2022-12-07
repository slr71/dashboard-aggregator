package db

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/lib/pq"
)

type AppsQueryConfig struct {
	Username          string
	GroupsIndex       int
	AppIDs            []string
	StartDateInterval string
}

func (d *Database) PopularFeaturedApps(ctx context.Context, cfg *AppsQueryConfig, opts ...QueryOption) ([]App, error) {
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

	a := goqu.T("app_listing")
	j := goqu.T("jobs")
	u := goqu.T("users")
	w := goqu.T("workspace")
	acg := goqu.T("app_category_group")
	aca := goqu.T("app_category_app")

	subquery := db.From(u).
		Join(w, goqu.On(u.Col("id").Eq(w.Col("user_id")))).
		Join(acg, goqu.On(w.Col("root_category_id").Eq(acg.Col("parent_category_id")))).
		Join(aca, goqu.On(acg.Col("child_category_id").Eq(aca.Col("app_category_id")))).
		Where(
			u.Col("username").Eq(cfg.Username),
			acg.Col("child_index").Eq(cfg.GroupsIndex),
			aca.Col("app_id").Eq(a.Col("id")),
		)

	query := db.From(a).
		Select(
			a.Col("id"),
			goqu.L(`'de'`).As("system_id"),
			a.Col("name"),
			a.Col("description"),
			a.Col("wiki_url"),
			a.Col("integration_date"),
			a.Col("edited_date"),
			a.Col("integrator_username").As(goqu.C("username")),
			goqu.COUNT(j.Col("id")).As(goqu.C("job_count")),
			goqu.L("EXISTS(?)", subquery).As(goqu.C("is_favorite")),
			goqu.L("true").As(goqu.C("is_public")),
		).
		LeftJoin(j, goqu.On(j.Col("app_id").Eq(goqu.Cast(a.Col("id"), "TEXT")))).
		Where(
			a.Col("id").Eq(goqu.Any(pq.Array(cfg.AppIDs))),
			a.Col("deleted").Eq(goqu.L("false")),
			a.Col("disabled").Eq(goqu.L("false")),
			a.Col("integration_date").IsNotNull(),
			goqu.Or(
				j.Col("start_date").Gte(goqu.L("now() - ?", goqu.Cast(goqu.L(fmt.Sprintf("'%s'", cfg.StartDateInterval)), "interval"))),
				j.Col("start_date").IsNull(),
			),
		).
		GroupBy(
			a.Col("id"),
			a.Col("name"),
			a.Col("description"),
			a.Col("wiki_url"),
			a.Col("integration_date"),
			a.Col("edited_date"),
			a.Col("integrator_username"),
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

	apps = make([]App, 0)
	if err = executor.ScanStructsContext(ctx, &apps); err != nil {
		return nil, err
	}

	return apps, err
}

func (d *Database) PublicAppsQuery(ctx context.Context, username string, groupIndex int, publicAppIDs []string, opts ...QueryOption) ([]App, error) {
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

	a := goqu.T("app_listing")
	w := goqu.T("workspace")
	acg := goqu.T("app_category_group")
	aca := goqu.T("app_category_app")
	u := goqu.T("users")

	subquery := db.From(u).
		Join(w, goqu.On(u.Col("id").Eq(w.Col("user_id")))).
		Join(acg, goqu.On(w.Col("root_category_id").Eq(acg.Col("parent_category_id")))).
		Join(aca, goqu.On(acg.Col("child_category_id").Eq(aca.Col("app_category_id")))).
		Where(
			u.Col("username").Eq(username),
			acg.Col("child_index").Eq(groupIndex),
			aca.Col("app_id").Eq(a.Col("id")),
		)

	query := db.From(a).
		Select(
			a.Col("id"),
			goqu.L(`'de'`).As(goqu.C("system_id")),
			a.Col("name"),
			a.Col("description"),
			a.Col("wiki_url"),
			a.Col("integration_date"),
			a.Col("edited_date"),
			a.Col("integrator_username").As(goqu.C("username")),
			goqu.L("EXISTS(?)", subquery).As(goqu.C("is_favorite")),
			goqu.L("true").As(goqu.C("is_public")),
		).
		Where(
			a.Col("id").Eq(goqu.Any(pq.Array(publicAppIDs))),
			a.Col("deleted").Eq(goqu.L("false")),
			a.Col("disabled").Eq(goqu.L("false")),
			a.Col("integration_date").IsNotNull(),
		).
		Order(
			a.Col("integration_date").Desc(),
		)

	if querySettings.hasLimit {
		query = query.Limit(querySettings.limit)
	}

	if querySettings.hasOffset {
		query = query.Offset(querySettings.offset)
	}

	executor := query.Executor()

	apps = make([]App, 0)
	if err = executor.ScanStructsContext(ctx, &apps); err != nil {
		return nil, err
	}

	return apps, nil
}

func (d *Database) RecentlyAddedApps(ctx context.Context, username string, groupIndex int, publicAppIDS []string, opts ...QueryOption) ([]App, error) {
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

	a := goqu.T("app_listing")
	w := goqu.T("workspace")
	acg := goqu.T("app_category_group")
	aca := goqu.T("app_category_app")
	u := goqu.T("users")

	subquery := db.From(u).
		Join(w, goqu.On(u.Col("id").Eq(w.Col("user_id")))).
		Join(acg, goqu.On(w.Col("root_category_id").Eq(acg.Col("parent_category_id")))).
		Join(aca, goqu.On(acg.Col("child_category_id").Eq(aca.Col("app_category_id")))).
		Where(
			u.Col("username").Eq(username),
			acg.Col("child_index").Eq(groupIndex),
			aca.Col("app_id").Eq(a.Col("id")),
		)

	query := db.From(a).
		Select(
			a.Col("id"),
			goqu.L(`'de'`).As(goqu.C("system_id")),
			a.Col("name"),
			a.Col("description"),
			a.Col("wiki_url"),
			a.Col("integration_date"),
			a.Col("edited_date"),
			a.Col("integrator_username").As(goqu.C("username")),
			goqu.L("EXISTS(?)", subquery).As(goqu.C("is_favorite")),
			a.Col("id").Eq(goqu.Any(pq.Array(publicAppIDS))).As(goqu.C("is_public")),
		).
		Where(
			a.Col("deleted").Eq(goqu.L("false")),
			a.Col("disabled").Eq(goqu.L("false")),
			a.Col("integrator_username").Eq(username),
		).
		Order(
			a.Col("integration_date").Desc(),
		)

	if querySettings.hasLimit {
		query = query.Limit(querySettings.limit)
	}

	if querySettings.hasOffset {
		query = query.Offset(querySettings.offset)
	}

	executor := query.Executor()

	apps = make([]App, 0)
	if err = executor.ScanStructsContext(ctx, &apps); err != nil {
		return nil, err
	}

	return apps, nil
}

func (d *Database) RecentlyUsedApps(ctx context.Context, cfg *AppsQueryConfig, opts ...QueryOption) ([]App, error) {
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

	a := goqu.T("app_listing")
	j := goqu.T("jobs")
	w := goqu.T("workspace")
	acg := goqu.T("app_category_group")
	aca := goqu.T("app_category_app")
	u := goqu.T("users")

	subquery := db.From(u).
		Join(w, goqu.On(u.Col("id").Eq(w.Col("user_id")))).
		Join(acg, goqu.On(w.Col("root_category_id").Eq(acg.Col("parent_category_id")))).
		Join(aca, goqu.On(acg.Col("child_category_id").Eq(aca.Col("app_category_id")))).
		Where(
			u.Col("username").Eq(cfg.Username),
			acg.Col("child_index").Eq(cfg.GroupsIndex),
			aca.Col("app_id").Eq(a.Col("id")),
		)

	query := db.From(j).
		Select(
			a.Col("id"),
			goqu.L(`'de'`).As(goqu.C("system_id")),
			a.Col("name"),
			a.Col("description"),
			a.Col("wiki_url"),
			a.Col("integration_date"),
			a.Col("edited_date"),
			a.Col("integrator_username").As(goqu.C("username")),
			goqu.L("EXISTS(?)", subquery),
			a.Col("id").Eq(goqu.Any(pq.Array(cfg.AppIDs))).As(goqu.C("is_public")),
			goqu.MAX(j.Col("start_date")).As(goqu.C("most_recent_start_date")),
		).
		Join(u, goqu.On(j.Col("user_id").Eq(u.Col("id")))).
		Join(a, goqu.On(goqu.Cast(a.Col("id"), "TEXT").Eq(j.Col("app_id")))).
		Where(
			u.Col("username").Eq(cfg.Username),
			a.Col("deleted").IsFalse(),
			a.Col("disabled").IsFalse(),
			j.Col("start_date").Gt(goqu.L("now() - ?", goqu.Cast(goqu.L(fmt.Sprintf("'%s'", cfg.StartDateInterval)), "INTERVAL"))),
		).
		GroupBy(
			a.Col("id"),
			a.Col("name"),
			a.Col("description"),
			a.Col("wiki_url"),
			a.Col("integration_date"),
			a.Col("edited_date"),
			a.Col("integrator_username"),
		).
		Order(
			goqu.C("most_recent_start_date").Desc(),
		)

	if querySettings.hasLimit {
		query = query.Limit(querySettings.limit)
	}

	if querySettings.hasOffset {
		query = query.Offset(querySettings.offset)
	}

	executor := query.Executor()

	apps = make([]App, 0)
	if err = executor.ScanStructsContext(ctx, &apps); err != nil {
		return nil, err
	}

	return apps, nil
}
