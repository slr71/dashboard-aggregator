package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strconv"

	_ "expvar"

	"github.com/cyverse-de/dashboard-aggregator/app"
	"github.com/cyverse-de/dashboard-aggregator/config"
	"github.com/cyverse-de/dashboard-aggregator/db"
	"github.com/cyverse-de/dashboard-aggregator/feeds"
	"github.com/cyverse-de/go-mod/cfg"
	"github.com/cyverse-de/go-mod/logging"
	"github.com/cyverse-de/go-mod/otelutils"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf"
	_ "github.com/lib/pq"
)

var log = logging.Log.WithField("package", "main")
var serviceName = "dashboard-aggregator"

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

func main() {
	var (
		err    error
		c      *koanf.Koanf
		dbconn *sqlx.DB

		cfgPath    = flag.String("config", cfg.DefaultConfigPath, "Path to the config file")
		dotEnvPath = flag.String("dotenv", cfg.DefaultDotEnvPath, "Path to the dotenv file")
		envPrefix  = flag.String("env-prefix", cfg.DefaultEnvPrefix, "The prefix for environment variables")
		itemLimit  = flag.Int("item-limit", 10, "The default limit on the number of items returned for a dashboard section")
		logLevel   = flag.String("log-level", "debug", "One of trace, debug, info, warn, error, fatal, or panic")
		listenPort = flag.Int("port", 3000, "The port the service listens on for requests")
	)

	flag.Parse()
	logging.SetupLogging(*logLevel)
	log := log.WithField("context", "main")

	var tracerCtx, cancel = context.WithCancel(context.Background())
	defer cancel()
	shutdown := otelutils.TracerProviderFromEnv(tracerCtx, serviceName, func(e error) { log.Fatal(e) })
	defer shutdown()

	c, err = cfg.Init(&cfg.Settings{
		EnvPrefix:   *envPrefix,
		ConfigPath:  *cfgPath,
		DotEnvPath:  *dotEnvPath,
		StrictMerge: false,
		FileType:    cfg.YAML,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Done reading config from %s", *cfgPath)

	config, err := config.New(c)
	if err != nil {
		log.Fatal(err)
	}

	if config.ListenPort != *listenPort {
		config.ListenPort = *listenPort
	}

	log.Infof("Connecting to the database at %s:%d/%s", config.DB.Host, config.DB.Port, config.DB.Name)
	dbconn, err = db.Connect(config.DB)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Done connecting to the database")

	ctx := context.Background()
	pf := feeds.NewPublicFeeds()
	pf.AddFeed(ctx, "news", feeds.NewWebsiteFeed(config.Feeds.NewsFeedURL, *itemLimit))
	pf.AddFeed(ctx, "events", feeds.NewWebsiteFeed(config.Feeds.EventsFeedURL, *itemLimit))
	pf.AddFeed(ctx, "videos", feeds.NewVideoFeed(config.Feeds.VideosURL, *itemLimit))

	log.Info("Pulling items from feeds")
	pf.PullItems(ctx)
	log.Info("Done pulling items from feeds")

	log.Infof("Scheduling feed refreshes")
	if err = pf.ScheduleRefreshes(ctx); err != nil {
		log.Error(err)
	}
	log.Info("Done scheduling feed refreshes")

	database := db.New(dbconn)
	a, err := app.New(database, pf, config)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Setting the public group ID")
	// Set the group ID
	err = a.SetPublicID(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Done setting the public group ID")

	ae := a.Echo()

	log.Info("Starting the server")
	srv := fmt.Sprintf(":%s", strconv.Itoa(config.ListenPort))
	log.Fatal(http.ListenAndServe(srv, ae))
}
