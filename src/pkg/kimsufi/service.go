package kimsufi

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ovh/go-ovh/ovh"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var (
	cacheExpiration = 5 * time.Minute
	cacheCleanup    = 10 * time.Minute
)

type Config struct {
	URL    string
	Logger *logrus.Logger
}

type MultiService struct {
	cache   *cache.Cache
	clients map[string]*ovh.Client
	logger  *logrus.Logger
	url     *url.URL
}

type Service struct {
	cache  *cache.Cache
	client *ovh.Client
	logger *logrus.Logger
	url    *url.URL
}

func NewService(config Config) (*MultiService, error) {
	c := cache.New(cacheExpiration, cacheCleanup)

	s := &MultiService{
		cache:   c,
		clients: make(map[string]*ovh.Client, 0),
		logger:  config.Logger,
	}

	for endpoint := range ovh.Endpoints {
		e := strings.Split(endpoint, "-")
		if len(e) >= 2 {
			if e[0] == "ovh" {
				client, err := ovh.NewClient(ovh.Endpoints[endpoint], "none", "none", "none")
				client.Logger = NewRequestLogger(config.Logger)
				//c, err := ovh.NewEndpointClient(ovh.KimsufiEU)
				if err != nil {
					log.Errorf("failed to create OVH client for %s: %v", endpoint, err)
					return nil, err
				}

				s.clients[endpoint] = client
			}
		}
	}

	return s, nil
}

func (m *MultiService) Endpoint(endpoint string) *Service {
	for e, client := range m.clients {
		if e == endpoint {
			return &Service{
				cache:  m.cache,
				client: client,
				logger: m.logger,
				url:    m.url,
			}
		}
	}

	return nil
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

// https://eu.api.ovh.com/v1/order.json
func (s *Service) GetOrderSchema() (*Order, error) {
	u, err := url.Parse("/order.json")
	q := u.Query()
	q.Set("format", "openapi3")
	u.RawQuery = q.Encode()

	var order *Order
	cacheEntry, found := s.cache.Get(u.String())
	if found {
		s.logger.Debugf("cache hit: %s", u.String())
		order = cacheEntry.(*Order)

	} else {
		s.logger.Debugf("cache miss: %s", u.String())
		err = s.client.GetUnAuth(u.String(), &order)
		if err != nil {
			return nil, err
		}
		s.cache.Set(u.String(), order, cache.DefaultExpiration)
	}

	return order, nil
}
