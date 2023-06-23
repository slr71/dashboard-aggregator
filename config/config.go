package config

import (
	"errors"
	"net/url"

	"github.com/cyverse-de/go-mod/logging"
	"github.com/knadh/koanf"
)

var log = logging.Log.WithField("package", "config")

func feedURL(base, path string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	result := u.JoinPath(path)
	return result.String(), nil
}

type DatabaseConfiguration struct {
	User     string
	Password string
	Host     string
	Port     int
	Name     string
}

func NewDatabaseConfiguration(config *koanf.Koanf) (*DatabaseConfiguration, error) {
	var dbConfig DatabaseConfiguration

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
	dbConfig.Port = config.Int("db.port")
	if dbConfig.Port == 0 {
		return nil, errors.New("db.port must be set in the configuration")
	}
	dbConfig.Name = config.String("db.database")
	if dbConfig.Name == "" {
		return nil, errors.New("db.database must be set in the configuration")
	}

	return &dbConfig, nil
}

type LoggingConfiguration struct {
	Level string
	Label string
}

func NewLoggingConfiguration(config *koanf.Koanf) *LoggingConfiguration {
	l := config.String("logging.level")
	if l == "" {
		l = "info"
	}
	lbl := config.String("logging.label")
	if lbl == "" {
		lbl = "dashboard-aggregator"
	}
	return &LoggingConfiguration{
		Level: l,
		Label: lbl,
	}
}

type MetadataConfiguration struct {
	URL                   string
	FeaturedAppsAttribute string
	FeaturedAppsValue     string
}

func NewMetadataConfiguration(config *koanf.Koanf) (*MetadataConfiguration, error) {
	u := config.String("metadata.url")
	if u == "" {
		u = "http://metadata"
	}
	a := config.String("metadata.featured_apps_attr")
	if a == "" {
		return nil, errors.New("metadata.featured_apps_attr must be set in the configuration")
	}
	v := config.String("metadata.featured_apps_value")
	if v == "" {
		return nil, errors.New("metadata.featured_apps_value must be set in the configuration")
	}
	return &MetadataConfiguration{
		URL:                   u,
		FeaturedAppsAttribute: a,
		FeaturedAppsValue:     v,
	}, nil
}

type AppExposerConfiguration struct {
	URL  string
	User string
}

func NewAppExposerConfiguration(config *koanf.Koanf) (*AppExposerConfiguration, error) {
	u := config.String("app-exposer.url")
	if u == "" {
		u = "http://app-exposer"
	}
	au := config.String("app-exposer.user")
	if au == "" {
		return nil, errors.New("app-exposer.user must be set in the configuration")
	}
	return &AppExposerConfiguration{
		URL:  u,
		User: au,
	}, nil
}

type FeedsConfiguration struct {
	WebsiteURL    string
	NewsFeedURL   string
	EventsFeedURL string
	VideosURL     string
}

func NewFeedsConfiguration(config *koanf.Koanf) (*FeedsConfiguration, error) {
	websiteBase := config.String("website.url")
	if websiteBase == "" {
		return nil, errors.New("website.url must be set in the configuration")
	}
	newsPath := config.String("website.feeds.news")
	if newsPath == "" {
		return nil, errors.New("website.feeds.news must be set in  the configuration")
	}
	eventsPath := config.String("website.feeds.events")
	if eventsPath == "" {
		return nil, errors.New("website.feeds.events must be set in the configuration")
	}
	videosURL := config.String("videos.url")
	if videosURL == "" {
		return nil, errors.New("videos.url must be set in the configuration")
	}
	newsURL, err := feedURL(websiteBase, newsPath)
	if err != nil {
		return nil, err
	}
	eventsURL, err := feedURL(websiteBase, eventsPath)
	if err != nil {
		return nil, err
	}
	return &FeedsConfiguration{
		WebsiteURL:    websiteBase,
		NewsFeedURL:   newsURL,
		EventsFeedURL: eventsURL,
		VideosURL:     videosURL,
	}, nil
}

type AppsConfiguration struct {
	URL                 string
	FavoritesGroupIndex int
}

func NewAppsConfiguration(config *koanf.Koanf) (*AppsConfiguration, error) {
	u := config.String("apps.url")
	if u == "" {
		return nil, errors.New("apps.url must be set in the configuration")
	}
	i := config.Int("apps.favorites_group_index")
	if i == 0 {
		i = 10
	}
	return &AppsConfiguration{
		URL:                 u,
		FavoritesGroupIndex: i,
	}, nil

}

type PermissionsConfiguration struct {
	GroupURL    string
	URL         string
	PublicGroup string
}

func NewPermissionsConfiguration(config *koanf.Koanf) (*PermissionsConfiguration, error) {
	u := config.String("permissions.uri")
	if u == "" {
		return nil, errors.New("permissions.uri must be set in the configuration")
	}

	g := config.String("permissions.public_group")
	if g == "" {
		return nil, errors.New("permissions.public_group must be set in the configuration")
	}
	log.Debug(g)

	i := config.String("iplant_groups.uri")
	if i == "" {
		return nil, errors.New("iplant_groups.uri must be set in the configuration")
	}
	log.Debug(i)

	return &PermissionsConfiguration{
		GroupURL:    i,
		URL:         u,
		PublicGroup: g,
	}, nil
}

// ServiceConfiguration is the type all other configuration types are included
// in.
type ServiceConfiguration struct {
	DB          *DatabaseConfiguration
	Logging     *LoggingConfiguration
	Feeds       *FeedsConfiguration
	AppExposer  *AppExposerConfiguration
	Metadata    *MetadataConfiguration
	Apps        *AppsConfiguration
	Permissions *PermissionsConfiguration
	ListenPort  int
}

func New(config *koanf.Koanf) (*ServiceConfiguration, error) {
	loggingConfig := NewLoggingConfiguration(config)
	dbConfig, err := NewDatabaseConfiguration(config)
	if err != nil {
		return nil, err
	}
	feedsConfig, err := NewFeedsConfiguration(config)
	if err != nil {
		return nil, err
	}
	appsConfig, err := NewAppsConfiguration(config)
	if err != nil {
		return nil, err
	}
	appExposerConfig, err := NewAppExposerConfiguration(config)
	if err != nil {
		return nil, err
	}
	mdConfig, err := NewMetadataConfiguration(config)
	if err != nil {
		return nil, err
	}
	permissionsConfig, err := NewPermissionsConfiguration(config)
	if err != nil {
		return nil, err
	}
	listenPort := config.Int("listen_port")
	if listenPort == 0 {
		listenPort = 60000
	}
	return &ServiceConfiguration{
		DB:          dbConfig,
		Logging:     loggingConfig,
		Feeds:       feedsConfig,
		AppExposer:  appExposerConfig,
		Metadata:    mdConfig,
		Apps:        appsConfig,
		Permissions: permissionsConfig,
		ListenPort:  listenPort,
	}, nil
}
