package authorizer

import (
	"fmt"
	"net/http"
	"time"

	"github.com/3scale/3scale-authorizer/pkg/backend/v1"
	"github.com/3scale/3scale-authorizer/pkg/core"
	"github.com/3scale/3scale-authorizer/pkg/system/v1/cache"
	"github.com/3scale/3scale-go-client/threescale"
	"github.com/3scale/3scale-go-client/threescale/api"
	"github.com/3scale/3scale-porta-go-client/client"
)

// Manager manages connections and interactions between the adapter and 3scale (system and backend)
// Supports managing interactions between multiple hosts and can optionally leverage available caching implementations
// Capable of Authorizing a request to 3scale and providing the required functionality to pull from the sources to do so
type Manager struct {
	clientBuilder  builder
	systemCache    *SystemCache
	backendConf    BackendConfig
	cachedBackends map[string]cachedBackend
	// stopFlush controls the background process that periodically flushes the cache
	stopFlush       chan struct{}
	metricsReporter *MetricsReporter
}

// SystemCache wraps the caching implementation and its configuration for 3scale system
type SystemCache struct {
	cache.ConfigurationCache
	SystemCacheConfig
	stopRefreshingTask chan struct{}
}

// SystemCacheConfig holds the configuration for the cache
type SystemCacheConfig struct {
	MaxSize               int
	NumRetryFailedRefresh int
	RefreshInterval       time.Duration
	TTL                   time.Duration
}

// SystemRequest provides the required input to request the latest configuration from 3scale system
type SystemRequest struct {
	AccessToken string
	ServiceID   string
	Environment string
}

type BackendConfig struct {
	// EnableCaching of authorization responses to 3scale
	EnableCaching bool
	// CacheFlushInterval is the period at which the cache should be flushed and
	// reported to 3scale
	CacheFlushInterval time.Duration
	Logger             core.Logger
	Policy             backend.FailurePolicy
}

// BackendAuth contains client authorization credentials for apisonator
type BackendAuth struct {
	Type  string
	Value string
}

// BackendRequest contains the data required to make an Auth/AuthRep request to apisonator
type BackendRequest struct {
	Auth         BackendAuth
	Service      string
	Transactions []BackendTransaction
}

// BackendResponse contains the result of an Auth/AuthRep request
type BackendResponse struct {
	Authorized bool
	ErrorCode  string
	// RejectedReason should* be set in cases where Authorized is false
	RejectedReason string
	RawResponse    interface{}
}

// BackendTransaction contains the metrics and end user auth required to make an Auth/AuthRep request to apisonator
type BackendTransaction struct {
	Metrics map[string]int
	Params  BackendParams
}

// BackendParams contains the ebd user auth for the various supported authentication patterns
type BackendParams struct {
	AppID   string
	AppKey  string
	UserID  string
	UserKey string
}

type cachedBackend struct {
	backend   *backend.Backend
	stopFlush chan struct{}
}

// NewManager returns an instance of Manager
// Starts refreshing background process for underlying system cache if provided
func NewManager(
	client *http.Client,
	systemCache *SystemCache,
	backendConfig BackendConfig,
	reporter *MetricsReporter,
) *Manager {

	builder := ClientBuilder{httpClient: http.DefaultClient}

	if reporter == nil {
		reporter = &MetricsReporter{}
	}

	if reporter.ReportMetrics && reporter.ResponseCB != nil {
		builder.httpClient.Transport = &MetricsTransport{client: builder.httpClient}
	}

	if systemCache != nil {
		go func() {
			ticker := time.NewTicker(systemCache.RefreshInterval)
			for {
				select {
				case <-ticker.C:
					systemCache.Refresh()
				case <-systemCache.stopRefreshingTask:
					ticker.Stop()
					return
				}
			}
		}()

	}

	m := &Manager{
		clientBuilder:   builder,
		systemCache:     systemCache,
		backendConf:     backendConfig,
		stopFlush:       make(chan struct{}),
		metricsReporter: reporter,
	}

	if backendConfig.EnableCaching {
		m.cachedBackends = make(map[string]cachedBackend)
	}

	return m
}

// NewSystemCache returns a system cache configured with an in-memory caching implementation
// and sets some sensible defaults if zero values have been provided for the config
func NewSystemCache(config SystemCacheConfig, stopRefreshing chan struct{}) *SystemCache {
	c := cache.NewConfigCache(config.TTL, config.MaxSize)

	if config.RefreshInterval == time.Duration(0) {
		config.RefreshInterval = cache.DefaultCacheRefreshInterval
	}

	if config.TTL == time.Duration(0) {
		config.TTL = cache.DefaultCacheTTL
	}

	return &SystemCache{
		ConfigurationCache: c,
		stopRefreshingTask: stopRefreshing,
		SystemCacheConfig:  config,
	}
}

// GetSystemConfiguration returns the configuration from 3scale system which can be used to fulfill and Auth request
func (m Manager) GetSystemConfiguration(systemURL string, request SystemRequest) (client.ProxyConfig, error) {
	var config client.ProxyConfig
	var err error

	if err = validateSystemRequest(request); err != nil {
		return config, err
	}

	if m.systemCache != nil && m.systemCache.ConfigurationCache != nil {
		config, err = m.fetchSystemConfigFromCache(systemURL, request)

	} else {
		config, err = m.fetchSystemConfigRemotely(systemURL, request)
	}

	if err != nil {
		return config, fmt.Errorf("cannot get 3scale system config - %s", err.Error())
	}

	return config, nil
}

// Shutdown stops running background process
func (m Manager) Shutdown() {
	close(m.stopFlush)
	close(m.systemCache.stopRefreshingTask)
}

// AuthRep does a Authorize and Report request into 3scale apisonator
func (m Manager) AuthRep(backendURL string, request BackendRequest) (*BackendResponse, error) {
	if !m.backendConf.EnableCaching {
		return m.passthroughAuthRep(backendURL, request)
	}

	return m.cachedAuthRep(backendURL, request)
}

func (m Manager) passthroughAuthRep(backendURL string, request BackendRequest) (*BackendResponse, error) {
	client, err := m.clientBuilder.BuildBackendClient(backendURL)
	if err != nil {
		return nil, fmt.Errorf("unable to build required client for 3scale backend - %s", err.Error())
	}

	return m.authRep(client, request)
}

func (m Manager) cachedAuthRep(backendURL string, request BackendRequest) (*BackendResponse, error) {
	var cb cachedBackend
	var err error
	cb, knownBackend := m.cachedBackends[backendURL]
	if !knownBackend {
		// try to create a cache if we haven't seen this backend before
		cb, err = m.newCachedBackend(backendURL)
		if err != nil {
			//todo(pgough) - add logging when we accept a logger
			return m.passthroughAuthRep(backendURL, request)
		}
		m.cachedBackends[backendURL] = cb
	}

	return m.authRep(cb.backend, request)
}

func (m Manager) authRep(client threescale.Client, request BackendRequest) (*BackendResponse, error) {
	req, err := request.ToAPIRequest()
	if err != nil {
		return nil, fmt.Errorf("unable to build request to 3scale - %s", err)
	}

	res, err := client.AuthRep(*req)
	if err != nil {
		var rawResponse interface{}
		if res != nil {
			rawResponse = res.RawResponse
		}
		return &BackendResponse{
			Authorized:  false,
			RawResponse: rawResponse,
		}, fmt.Errorf("error calling AuthRep - %s", err)
	}

	return &BackendResponse{
		Authorized:     res.Authorized,
		ErrorCode:      res.ErrorCode,
		RejectedReason: res.RejectionReason,
		RawResponse:    res.RawResponse,
	}, nil
}

// newCachedBackend creates a new backend and start the flushing process in the background
func (m Manager) newCachedBackend(url string) (cachedBackend, error) {
	httpClient := http.DefaultClient
	if cb, ok := m.clientBuilder.(*ClientBuilder); ok {
		httpClient = cb.httpClient
	}
	backend, err := backend.NewBackend(url, httpClient, m.backendConf.Logger, m.backendConf.Policy)
	if err != nil {
		return cachedBackend{}, err
	}

	ticker := time.NewTicker(m.backendConf.CacheFlushInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				backend.Flush()
			case <-m.stopFlush:
				// allows us to drain the cache before shutting down
				backend.Flush()
				ticker.Stop()
				return
			}

		}
	}()
	m.backendConf.Logger.Infof("created new cached backend for %s", url)
	return cachedBackend{
		backend:   backend,
		stopFlush: m.stopFlush,
	}, nil
}

func (m Manager) fetchSystemConfigFromCache(systemURL string, request SystemRequest) (client.ProxyConfig, error) {
	var config client.ProxyConfig
	var err error

	cacheKey := generateSystemCacheKey(systemURL, request.ServiceID)
	cachedValue, found := m.systemCache.Get(cacheKey)
	if !found {
		config, err = m.fetchSystemConfigRemotely(systemURL, request)
		if err != nil {
			return config, err
		}

		itemToCache := &cache.Value{Item: config}
		itemToCache = m.setValueFromConfig(systemURL, request, itemToCache)
		m.systemCache.Set(cacheKey, *itemToCache)

	} else {
		config = cachedValue.Item
		if m.metricsReporter.CacheHitCB != nil {
			m.metricsReporter.CacheHitCB(System)
		}
	}

	return config, err
}

func (m Manager) fetchSystemConfigRemotely(systemURL string, request SystemRequest) (client.ProxyConfig, error) {
	var config client.ProxyConfig

	systemClient, err := m.clientBuilder.BuildSystemClient(systemURL, request.AccessToken)
	if err != nil {
		return config, fmt.Errorf("unable to build system client for %s - %s", systemURL, err.Error())
	}

	proxyConfElement, err := systemClient.GetLatestProxyConfig(request.ServiceID, request.Environment)
	if err != nil {
		return config, fmt.Errorf("unable to fetch required data from 3scale system - %s", err.Error())
	}

	return proxyConfElement.ProxyConfig, nil
}

func (m Manager) refreshCallback(systemURL string, request SystemRequest, retryAttempts int) func() (client.ProxyConfig, error) {
	return func() (client.ProxyConfig, error) {
		config, err := m.fetchSystemConfigRemotely(systemURL, request)
		if err != nil {
			if retryAttempts > 0 {
				retryAttempts--
				return m.refreshCallback(systemURL, request, retryAttempts)()
			}
		}
		return config, err
	}
}

func (m Manager) setValueFromConfig(systemURL string, request SystemRequest, value *cache.Value) *cache.Value {
	value.SetRefreshCallback(m.refreshCallback(systemURL, request, m.systemCache.NumRetryFailedRefresh))
	return value
}

// ToAPIRequest transforms the BackendRequest into a request that is acceptable for the 3scale Client interface
func (request BackendRequest) ToAPIRequest() (*threescale.Request, error) {
	if request.Transactions == nil || len(request.Transactions) < 1 {
		return nil, fmt.Errorf("cannot process emtpy transaction")
	}

	return &threescale.Request{
		Auth: api.ClientAuth{
			Type:  api.AuthType(request.Auth.Type),
			Value: request.Auth.Value,
		},
		// we want to be have 3scale set the error_code explicitly
		Extensions: api.Extensions{
			backend.RejectionReasonHeaderExtension: "1",
		},
		Service: api.Service(request.Service),
		Transactions: []api.Transaction{
			{
				Metrics: request.Transactions[0].Metrics,
				Params: api.Params{
					AppID:   request.Transactions[0].Params.AppID,
					AppKey:  request.Transactions[0].Params.AppKey,
					UserID:  request.Transactions[0].Params.UserID,
					UserKey: request.Transactions[0].Params.UserKey,
				},
			},
		},
	}, nil
}

// validateSystemRequest to avoid wasting compute time on invalid request
func validateSystemRequest(request SystemRequest) error {
	if request.Environment == "" || request.ServiceID == "" || request.AccessToken == "" {
		return fmt.Errorf("invalid arguements provided")
	}
	return nil
}

func generateSystemCacheKey(systemURL, svcID string) string {
	return fmt.Sprintf("%s_%s", systemURL, svcID)
}
