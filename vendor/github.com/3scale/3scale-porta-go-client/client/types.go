package client

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"time"
)

// AdminPortal defines a 3scale adminPortal service
type AdminPortal struct {
	scheme  string
	host    string
	port    int
	baseUrl *url.URL
}

// ThreeScaleClient interacts with 3scale Service Management API
type ThreeScaleClient struct {
	adminPortal *AdminPortal
	credential  string
	httpClient  *http.Client
}

// Application - API response for create app endpoint
type Application struct {
	ID                      int64  `json:"id"`
	CreatedAt               string `json:"created_at"`
	UpdatedAt               string `json:"updated_at"`
	State                   string `json:"state"`
	UserAccountID           string `json:"user_account_id"`
	FirstTrafficAt          string `json:"first_traffic_at"`
	FirstDailyTrafficAt     string `json:"first_daily_traffic_at"`
	EndUserRequired         bool   `json:"end_user_required"`
	ServiceID               int64  `json:"service_id"`
	UserKey                 string `json:"user_key"`
	ProviderVerificationKey string `json:"provider_verification_key"`
	PlanID                  int64  `json:"plan_id"`
	AppName                 string `json:"name"`
	Description             string `json:"description"`
	ExtraFields             string `json:"extra_fields"`
	Error                   string `json:"error,omitempty"`
}

// ApplicationElem - Holds a intenal application element
type ApplicationElem struct {
	Application Application `json:"application"`
}

// ApplicationList - Holds a list of applications
type ApplicationList struct {
	Applications []ApplicationElem `json:"applications"`
}

// ApplicationPlansList - Holds a list of application plans
type ApplicationPlansList struct {
	XMLName xml.Name `xml:"plans"`
	Plans   []Plan   `xml:"plan"`
}

// Limit - Defines the object returned via the API for creation of a limit
type Limit struct {
	XMLName  xml.Name `xml:"limit"`
	ID       string   `xml:"id"`
	MetricID string   `xml:"metric_id"`
	PlanID   string   `xml:"plan_id"`
	Period   string   `xml:"period"`
	Value    string   `xml:"value"`
}

// LimitList - Holds a list of Limit
type LimitList struct {
	XMLName xml.Name `xml:"limits"`
	Limits  []Limit  `xml:"limit"`
}

// MappingRule - Defines the object returned via the API for creation of mapping rule
type MappingRule struct {
	XMLName    xml.Name `xml:"mapping_rule"`
	ID         string   `xml:"id,omitempty"`
	MetricID   string   `xml:"metric_id,omitempty"`
	Pattern    string   `xml:"pattern,omitempty"`
	HTTPMethod string   `xml:"http_method,omitempty"`
	Delta      string   `xml:"delta,omitempty"`
	CreatedAt  string   `xml:"created_at,omitempty"`
	UpdatedAt  string   `xml:"updated_at,omitempty"`
}

// MappingRuleList - Holds a list of MappingRule
type MappingRuleList struct {
	XMLName      xml.Name      `xml:"mapping_rules"`
	MappingRules []MappingRule `xml:"mapping_rule"`
}

// Metric - Defines the object returned via the API for creation of metric
type Metric struct {
	XMLName      xml.Name `xml:"metric"`
	ID           string   `xml:"id"`
	MetricName   string   `xml:"name"`
	SystemName   string   `xml:"system_name"`
	FriendlyName string   `xml:"friendly_name"`
	ServiceID    string   `xml:"service_id"`
	Description  string   `xml:"description"`
	Unit         string   `xml:"unit"`
}

// MetricList - Holds a list of Metric
type MetricList struct {
	XMLName xml.Name `xml:"metrics"`
	Metrics []Metric `xml:"metric"`
}

// Plan - API response for create application plan endpoint
type Plan struct {
	XMLNameName        xml.Name `xml:"plan"`
	Custom             string   `xml:"custom,attr"`
	Default            bool     `xml:"default,attr"`
	ID                 string   `xml:"id"`
	PlanName           string   `xml:"name"`
	Type               string   `xml:"type"`
	State              string   `xml:"state"`
	ServiceID          string   `xml:"service_id"`
	EndUserRequired    string   `xml:"end_user_required"`
	ApprovalRequired   string   `xml:"approval_required"`
	SetupFee           string   `xml:"setup_fee"`
	CostPerMonth       string   `xml:"cost_per_month"`
	TrialPeriodDays    string   `xml:"trial_period_days"`
	CancellationPeriod string   `xml:"cancellation_period"`
	Error              string   `xml:"error,omitempty"`
}

type Proxy struct {
	XMLName                 xml.Name `xml:"proxy"`
	ServiceID               string   `xml:"service_id"`
	Endpoint                string   `xml:"endpoint"`
	ApiBackend              string   `xml:"api_backend"`
	CredentialsLocation     string   `xml:"credentials_location"`
	AuthAppKey              string   `xml:"auth_app_key"`
	AuthAppID               string   `xml:"auth_app_id"`
	AuthUserKey             string   `xml:"auth_user_key"`
	ErrorAuthFailed         string   `xml:"error_auth_failed"`
	ErrorAuthMissing        string   `xml:"error_auth_missing"`
	ErrorStatusAuthFailed   string   `xml:"error_status_auth_failed"`
	ErrorHeadersAuthFailed  string   `xml:"error_headers_auth_failed"`
	ErrorStatusAuthMissing  string   `xml:"error_status_auth_missing"`
	ErrorHeadersAuthMissing string   `xml:"error_headers_auth_missing"`
	ErrorNoMatch            string   `xml:"error_no_match"`
	ErrorStatusNoMatch      string   `xml:"error_status_no_match"`
	ErrorHeadersNoMatch     string   `xml:"error_headers_no_match"`
	SecretToken             string   `xml:"secret_token"`
	HostnameRewrite         string   `xml:"hostname_rewrite"`
	SandboxEndpoint         string   `xml:"sandbox_endpoint"`
	ApiTestPath             string   `xml:"api_test_path"`
	PoliciesConfig          string   `xml:"policies_config"`
	CreatedAt               string   `xml:"created_at"`
	UpdatedAt               string   `xml:"updated_at"`
	LockVersion             string   `xml:"lock_version"`
	OidcIssuerEndpoint      string   `xml:"oidc_issuer_endpoint"`
}

type Service struct {
	ID                          string     `xml:"id"`
	AccountID                   string     `xml:"account_id"`
	Name                        string     `xml:"name"`
	Description                 string     `xml:"description"`
	DeploymentOption            string     `xml:"deployment_option"`
	State                       string     `xml:"state"`
	SystemName                  string     `xml:"system_name"`
	BackendVersion              string     `xml:"backend_version"`
	EndUserRegistrationRequired string     `xml:"end_user_registration_required"`
	Metrics                     MetricList `xml:"metrics"`
}

type ServiceList struct {
	XMLName  xml.Name  `xml:"services"`
	Services []Service `xml:"service"`
}

type ErrorResp struct {
	XMLName xml.Name `xml:"error"`
	Text    string   `xml:",chardata"`
	Error   struct {
		Text string `xml:",chardata"`
	} `xml:"error"`
}

// Following structs with JSON tags are used in the Proxy Config APIs which return JSON

type ProxyConfig struct {
	ID          int     `json:"id"`
	Version     int     `json:"version"`
	Environment string  `json:"environment"`
	Content     Content `json:"content"`
}

type ProxyConfigList struct {
	ProxyConfigs []ProxyConfigElement `json:"proxy_configs"`
}

type ProxyConfigElement struct {
	ProxyConfig ProxyConfig `json:"proxy_config"`
}

type Content struct {
	ID                          int64        `json:"id"`
	AccountID                   int64        `json:"account_id"`
	Name                        string       `json:"name"`
	OnelineDescription          interface{}  `json:"oneline_description"`
	Description                 interface{}  `json:"description"`
	TxtAPI                      interface{}  `json:"txt_api"`
	TxtSupport                  interface{}  `json:"txt_support"`
	TxtFeatures                 interface{}  `json:"txt_features"`
	CreatedAt                   time.Time    `json:"created_at"`
	UpdatedAt                   time.Time    `json:"updated_at"`
	LogoFileName                interface{}  `json:"logo_file_name"`
	LogoContentType             interface{}  `json:"logo_content_type"`
	LogoFileSize                interface{}  `json:"logo_file_size"`
	State                       string       `json:"state"`
	IntentionsRequired          bool         `json:"intentions_required"`
	DraftName                   string       `json:"draft_name"`
	Infobar                     interface{}  `json:"infobar"`
	Terms                       interface{}  `json:"terms"`
	DisplayProviderKeys         bool         `json:"display_provider_keys"`
	TechSupportEmail            interface{}  `json:"tech_support_email"`
	AdminSupportEmail           interface{}  `json:"admin_support_email"`
	CreditCardSupportEmail      interface{}  `json:"credit_card_support_email"`
	BuyersManageApps            bool         `json:"buyers_manage_apps"`
	BuyersManageKeys            bool         `json:"buyers_manage_keys"`
	CustomKeysEnabled           bool         `json:"custom_keys_enabled"`
	BuyerPlanChangePermission   string       `json:"buyer_plan_change_permission"`
	BuyerCanSelectPlan          bool         `json:"buyer_can_select_plan"`
	NotificationSettings        interface{}  `json:"notification_settings"`
	DefaultApplicationPlanID    int64        `json:"default_application_plan_id"`
	DefaultServicePlanID        int64        `json:"default_service_plan_id"`
	DefaultEndUserPlanID        interface{}  `json:"default_end_user_plan_id"`
	EndUserRegistrationRequired bool         `json:"end_user_registration_required"`
	TenantID                    int64        `json:"tenant_id"`
	SystemName                  string       `json:"system_name"`
	BackendVersion              string       `json:"backend_version"`
	MandatoryAppKey             bool         `json:"mandatory_app_key"`
	BuyerKeyRegenerateEnabled   bool         `json:"buyer_key_regenerate_enabled"`
	SupportEmail                string       `json:"support_email"`
	ReferrerFiltersRequired     bool         `json:"referrer_filters_required"`
	DeploymentOption            string       `json:"deployment_option"`
	Proxiable                   bool         `json:"proxiable?"`
	BackendAuthenticationType   string       `json:"backend_authentication_type"`
	BackendAuthenticationValue  string       `json:"backend_authentication_value"`
	Proxy                       ContentProxy `json:"proxy"`
}

type ContentProxy struct {
	ID                         int64         `json:"id"`
	TenantID                   int64         `json:"tenant_id"`
	ServiceID                  int64         `json:"service_id"`
	Endpoint                   string        `json:"endpoint"`
	DeployedAt                 interface{}   `json:"deployed_at"`
	APIBackend                 string        `json:"api_backend"`
	AuthAppKey                 string        `json:"auth_app_key"`
	AuthAppID                  string        `json:"auth_app_id"`
	AuthUserKey                string        `json:"auth_user_key"`
	CredentialsLocation        string        `json:"credentials_location"`
	ErrorAuthFailed            string        `json:"error_auth_failed"`
	ErrorAuthMissing           string        `json:"error_auth_missing"`
	CreatedAt                  string        `json:"created_at"`
	UpdatedAt                  string        `json:"updated_at"`
	ErrorStatusAuthFailed      int64         `json:"error_status_auth_failed"`
	ErrorHeadersAuthFailed     string        `json:"error_headers_auth_failed"`
	ErrorStatusAuthMissing     int64         `json:"error_status_auth_missing"`
	ErrorHeadersAuthMissing    string        `json:"error_headers_auth_missing"`
	ErrorNoMatch               string        `json:"error_no_match"`
	ErrorStatusNoMatch         int64         `json:"error_status_no_match"`
	ErrorHeadersNoMatch        string        `json:"error_headers_no_match"`
	SecretToken                string        `json:"secret_token"`
	HostnameRewrite            *string       `json:"hostname_rewrite"`
	OauthLoginURL              interface{}   `json:"oauth_login_url"`
	SandboxEndpoint            string        `json:"sandbox_endpoint"`
	APITestPath                string        `json:"api_test_path"`
	APITestSuccess             *bool         `json:"api_test_success"`
	ApicastConfigurationDriven bool          `json:"apicast_configuration_driven"`
	OidcIssuerEndpoint         interface{}   `json:"oidc_issuer_endpoint"`
	LockVersion                int64         `json:"lock_version"`
	AuthenticationMethod       string        `json:"authentication_method"`
	HostnameRewriteForSandbox  string        `json:"hostname_rewrite_for_sandbox"`
	EndpointPort               int64         `json:"endpoint_port"`
	Valid                      bool          `json:"valid?"`
	ServiceBackendVersion      string        `json:"service_backend_version"`
	Hosts                      []string      `json:"hosts"`
	Backend                    Backend       `json:"backend"`
	PolicyChain                []PolicyChain `json:"policy_chain"`
	ProxyRules                 []ProxyRule   `json:"proxy_rules"`
}

type Backend struct {
	Endpoint string `json:"endpoint"`
	Host     string `json:"host"`
}

type PolicyChain struct {
	Name          string        `json:"name"`
	Version       string        `json:"version"`
	Configuration Configuration `json:"configuration"`
}

type Configuration struct {
}

type ProxyRule struct {
	ID                    int64         `json:"id"`
	ProxyID               int64         `json:"proxy_id"`
	HTTPMethod            string        `json:"http_method"`
	Pattern               string        `json:"pattern"`
	MetricID              int64         `json:"metric_id"`
	MetricSystemName      string        `json:"metric_system_name"`
	Delta                 int64         `json:"delta"`
	TenantID              int64         `json:"tenant_id"`
	CreatedAt             string        `json:"created_at"`
	UpdatedAt             string        `json:"updated_at"`
	RedirectURL           interface{}   `json:"redirect_url"`
	Parameters            []string      `json:"parameters"`
	QuerystringParameters Configuration `json:"querystring_parameters"`
}

type Params map[string]string

type User struct {
	ID        int64  `json:"id"`
	State     string `json:"state"`
	UserName  string `json:"username"`
	Email     string `json:"email"`
	AccountID int64  `json:"account_id"`
}

type UserElem struct {
	User User `json:"user"`
}

type UserList struct {
	Users []UserElem `json:"users"`
}

type Account struct {
	ID           int64  `json:"id"`
	State        string `json:"state"`
	OrgName      string `json:"org_name"`
	SupportEmail string `json:"support_email"`
	AdminDomain  string `json:"admin_domain"`
	Domain       string `json:"domain"`
}

type AccountElem struct {
	Account Account `json:"account"`
}

type AccountList struct {
	Accounts []AccountElem `json:"accounts"`
}

type AccessToken struct {
	ID         int64    `json:"id"`
	Name       string   `json:"name"`
	Scopes     []string `json:"scopes"`
	Permission string   `json:"permission"`
	Value      string   `json:"value"`
}

type Signup struct {
	Account     Account     `json:"account"`
	AccessToken AccessToken `json:"access_token"`
}

type Tenant struct {
	Signup Signup `json:"signup"`
}