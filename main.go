package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	_ "expvar"

	"github.com/cyverse-de/dashboard-aggregator/app"
	"github.com/cyverse-de/dashboard-aggregator/db"
	"github.com/cyverse-de/dashboard-aggregator/feeds"
	"github.com/cyverse-de/go-mod/cfg"
	"github.com/cyverse-de/go-mod/logging"
	"github.com/cyverse-de/go-mod/otelutils"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf"
	_ "github.com/lib/pq"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	"github.com/uptrace/opentelemetry-go-extra/otelsqlx"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var log = logging.Log.WithField("package", "main")

const serviceName = "dashboard-aggregator"

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

func NewDB(config *koanf.Koanf) (*sqlx.DB, error) {
	var (
		dbConfig DBConfig
		dbconn   *sqlx.DB
	)
	dbConfig.User = config.String("db.user")
	if dbConfig.User == "" {
		return nil, errors.New("db.user must be set in the configuration file")
	}
	dbConfig.Password = config.String("db.password")
	if dbConfig.Password == "" {
		return nil, errors.New("db.password must be set in the configuration")
	}
	dbConfig.Host = config.String("db.host")
	if dbConfig.Host == "" {
		return nil, errors.New("db.host must be set in the configuration")
	}
	dbConfig.Port = config.String("db.port")
	if dbConfig.Port == "" {
		return nil, errors.New("db.port must be set in the configuration")
	}
	dbConfig.Name = config.String("db.database")
	if dbConfig.Name == "" {
		return nil, errors.New("db.database must be set in the configuration")
	}
	dbURI := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
	)
	dbconn = otelsqlx.MustConnect("postgres", dbURI,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	log.Info("done connecting to the database")
	dbconn.SetMaxOpenConns(10)
	dbconn.SetConnMaxIdleTime(time.Minute)
	return dbconn, nil
}

func feedURL(base, path string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	result := u.JoinPath(path)
	return result.String(), nil
}

func main() {
	var (
		err    error
		config *koanf.Koanf
		dbconn *sqlx.DB

		cfgPath    = flag.String("config", cfg.DefaultConfigPath, "Path to the config file")
		dotEnvPath = flag.String("dotenv", cfg.DefaultDotEnvPath, "Path to the dotenv file")
		envPrefix  = flag.String("env-prefix", cfg.DefaultEnvPrefix, "The prefix for environment variables")
		itemLimit  = flag.Int("item-limit", 10, "The default limit on the number of items returned for a dashboard section")
		logLevel   = flag.String("log-level", "debug", "One of trace, debug, info, warn, error, fatal, or panic")
		listenPort = flag.Int("port", 60000, "The port the service listens on for requests")
	)

	flag.Parse()
	logging.SetupLogging(*logLevel)
	log := log.WithField("context", "main")

	var tracerCtx, cancel = context.WithCancel(context.Background())
	defer cancel()
	shutdown := otelutils.TracerProviderFromEnv(tracerCtx, serviceName, func(e error) { log.Fatal(e) })
	defer shutdown()

	config, err = cfg.Init(&cfg.Settings{
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

	dbconn, err = NewDB(config)
	if err != nil {
		log.Fatal(err)
	}

	websiteBase := config.String("website.url")
	if websiteBase == "" {
		log.Fatal("website.url must be set in the configuration")
	}
	newsPath := config.String("website.feeds.news")
	if newsPath == "" {
		log.Fatal("website.feeds.news must be set in  the configuration")
	}
	eventsPath := config.String("website.feeds.events")
	if eventsPath == "" {
		log.Fatal("website.feeds.events must be set in the configuration")
	}
	videosURL := config.String("videos.url")
	if videosURL == "" {
		log.Fatal("videos.url must be set in the configuration")
	}
	newsURL, err := feedURL(websiteBase, newsPath)
	if err != nil {
		log.Fatal(err)
	}
	eventsURL, err := feedURL(websiteBase, eventsPath)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pf := feeds.NewPublicFeeds()
	pf.AddFeed(ctx, "news", feeds.NewWebsiteFeed(newsURL, *itemLimit))
	pf.AddFeed(ctx, "events", feeds.NewWebsiteFeed(eventsURL, *itemLimit))
	pf.AddFeed(ctx, "videos", feeds.NewVideoFeed(videosURL, *itemLimit))

	database := db.New(dbconn)
	a := app.New(database, pf).Echo()

	srv := fmt.Sprintf(":%s", strconv.Itoa(*listenPort))
	log.Fatal(http.ListenAndServe(srv, a))
}
