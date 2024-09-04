package kimsufi

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ovh/go-ovh/ovh"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

var (
	cacheExpiration = 5 * time.Minute
	cacheCleanup    = 10 * time.Minute
)

type Config struct {
	URL    string
	Logger *logrus.Logger
}

type Service struct {
	cache  *cache.Cache
	client *ovh.Client
	logger *logrus.Logger
	url    *url.URL
}

func NewService(config Config) (*Service, error) {
	client, err := ovh.NewClient(config.URL, "none", "none", "none")
	//c, err := ovh.NewEndpointClient(ovh.KimsufiEU)
	if err != nil {
		fmt.Println("nope")
		return nil, err
	}
	client.Logger = NewRequestLogger(config.Logger)

	c := cache.New(cacheExpiration, cacheCleanup)

	s := &Service{
		cache:  c,
		client: client,
		logger: config.Logger,
	}

	return s, nil
}

func (s *Service) GetAvailabilities(datacenters []string, planCode string) (*Availabilities, error) {
	u, err := url.Parse("/dedicated/server/datacenter/availabilities")
	q := u.Query()
	if len(datacenters) > 0 {
		q.Set("datacenters", strings.Join(datacenters, ","))
	}
	if planCode != "" {
		q.Set("planCode", planCode)
	}
	u.RawQuery = q.Encode()

	var availabilities *Availabilities
	cacheEntry, found := s.cache.Get(u.String())
	if found {
		s.logger.Debugf("cache hit: %s", u.String())
		availabilities = cacheEntry.(*Availabilities)
	} else {
		s.logger.Debugf("cache miss: %s", u.String())
		err = s.client.GetUnAuth(u.String(), &availabilities)
		if err != nil {
			return nil, err
		}
		s.cache.Set(u.String(), availabilities, cache.DefaultExpiration)
	}

	return availabilities, nil
}

func (s *Service) ListServers(ovhSubsidiary string) (*Catalog, error) {
	u, err := url.Parse("/order/catalog/public/eco")
	q := u.Query()
	q.Set("ovhSubsidiary", ovhSubsidiary)
	u.RawQuery = q.Encode()

	var catalog *Catalog
	cacheEntry, found := s.cache.Get(u.String())
	if found {
		s.logger.Debugf("cache hit: %s", u.String())
		catalog = cacheEntry.(*Catalog)

	} else {
		s.logger.Debugf("cache miss: %s", u.String())
		err = s.client.GetUnAuth(u.String(), &catalog)
		if err != nil {
			return nil, err
		}
		s.cache.Set(u.String(), catalog, cache.DefaultExpiration)
	}

	return catalog, nil
}

type Logger struct {
	logger *logrus.Logger
}

func NewRequestLogger(l *logrus.Logger) *Logger {
	logger := &Logger{
		logger: l,
	}

	return logger
}

func (l *Logger) LogRequest(r *http.Request) {
	l.logger.Debugf("kimsufi: %s %s %s %v\n", r.Method, r.URL.String(), r.Proto, r.Header)
}

func (l *Logger) LogResponse(r *http.Response) {
	l.logger.Debugf("kimsufi: %s %s %v\n", r.Status, r.Proto, r.Header)
}
