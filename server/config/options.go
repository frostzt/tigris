// Copyright 2022-2023 Tigris Data, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"time"

	"github.com/tigrisdata/tigris/util/log"
)

const (
	DatabaseServerType = "database"
	RealtimeServerType = "realtime"
)

type ServerConfig struct {
	Host         string
	Port         int16
	Type         string `mapstructure:"type" yaml:"type" json:"type"`
	FDBHardDrop  bool   `mapstructure:"fdb_hard_drop" yaml:"fdb_hard_drop" json:"fdb_hard_drop"`
	RealtimePort int16  `mapstructure:"realtime_port" yaml:"realtime_port" json:"realtime_port"`
}

type Config struct {
	Log           log.LogConfig
	Server        ServerConfig    `yaml:"server" json:"server"`
	Auth          AuthConfig      `yaml:"auth" json:"auth"`
	Cdc           CdcConfig       `yaml:"cdc" json:"cdc"`
	Search        SearchConfig    `yaml:"search" json:"search"`
	Cache         CacheConfig     `yaml:"cache" json:"cache"`
	Tracing       TracingConfig   `yaml:"tracing" json:"tracing"`
	Metrics       MetricsConfig   `yaml:"metrics" json:"metrics"`
	Profiling     ProfilingConfig `yaml:"profiling" json:"profiling"`
	FoundationDB  FoundationDBConfig
	Quota         QuotaConfig
	Observability ObservabilityConfig `yaml:"observability" json:"observability"`
	Management    ManagementConfig    `yaml:"management" json:"management"`
	Schema        SchemaConfig
}

type AuthConfig struct {
	Enabled                   bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	IssuerURL                 string        `mapstructure:"issuer_url" yaml:"issuer_url" json:"issuer_url"`
	Audience                  string        `mapstructure:"audience" yaml:"audience" json:"audience"`
	JWKSCacheTimeout          time.Duration `mapstructure:"jwks_cache_timeout" yaml:"jwks_cache_timeout" json:"jwks_cache_timeout"`
	LogOnly                   bool          `mapstructure:"log_only" yaml:"log_only" json:"log_only"`
	EnableNamespaceIsolation  bool          `mapstructure:"enable_namespace_isolation" yaml:"enable_namespace_isolation" json:"enable_namespace_isolation"`
	AdminNamespaces           []string      `mapstructure:"admin_namespaces" yaml:"admin_namespaces" json:"admin_namespaces"`
	OAuthProvider             string        `mapstructure:"oauth_provider" yaml:"oauth_provider" json:"oauth_provider"`
	ClientId                  string        `mapstructure:"client_id" yaml:"client_id" json:"client_id"`
	ExternalTokenURL          string        `mapstructure:"external_token_url" yaml:"external_token_url" json:"external_token_url"`
	EnableOauth               bool          `mapstructure:"enable_oauth" yaml:"enable_oauth" json:"enable_oauth"`
	TokenCacheSize            int           `mapstructure:"token_cache_size" yaml:"token_cache_size" json:"token_cache_size"`
	ExternalDomain            string        `mapstructure:"external_domain" yaml:"external_domain" json:"external_domain"`
	ManagementClientId        string        `mapstructure:"management_client_id" yaml:"management_client_id" json:"management_client_id"`
	ManagementClientSecret    string        `mapstructure:"management_client_secret" yaml:"management_client_secret" json:"management_client_secret"`
	TokenClockSkewDurationSec int           `mapstructure:"token_clock_skew_duration_sec" yaml:"token_clock_skew_duration_sec" json:"token_clock_skew_duration_sec"`
}

type CdcConfig struct {
	Enabled        bool `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	StreamInterval time.Duration
	StreamBatch    int
	StreamBuffer   int
}

type TracingConfig struct {
	Enabled bool                 `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Datadog DatadogTracingConfig `mapstructure:"datadog" yaml:"datadog" json:"datadog"`
	Jaeger  JaegerTracingConfig  `mapstructure:"jaeger" yaml:"jaeger" json:"jaeger"`
}

type DatadogTracingConfig struct {
	Enabled             bool    `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	SampleRate          float64 `mapstructure:"sample_rate" yaml:"sample_rate" json:"sample_rate"`
	CodeHotspotsEnabled bool    `mapstructure:"codehotspots_enabled" yaml:"codehotspots_enabled" json:"codehotspots_enabled"`
	EndpointsEnabled    bool    `mapstructure:"endpoints_enabled" yaml:"endpoints_enabled" json:"endpoints_enabled"`
	WithUDS             string  `mapstructure:"agent_socket" yaml:"agent_socket" json:"agent_socket"`
	WithAgentAddr       string  `mapstructure:"agent_addr" yaml:"agent_addr" json:"agent_addr"`
	WithDogStatsdAddr   string  `mapstructure:"dogstatsd_addr" yaml:"dogstatsd_addr" json:"dogstatsd_addr"`
}

type JaegerTracingConfig struct {
	Enabled    bool    `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Url        string  `mapstructure:"url" yaml:"url" json:"url"`
	SampleRate float64 `mapstructure:"sample_rate" yaml:"sample_rate" json:"sample_rate"`
}

type MetricsConfig struct {
	Enabled        bool                      `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	TimerQuantiles []float64                 `mapstructure:"quantiles" yaml:"quantiles" json:"quantiles"`
	Requests       RequestsMetricGroupConfig `mapstructure:"requests" yaml:"requests" json:"requests"`
	Fdb            FdbMetricGroupConfig      `mapstructure:"fdb" yaml:"fdb" json:"fdb"`
	Search         SearchMetricGroupConfig   `mapstructure:"search" yaml:"search" json:"search"`
	Session        SessionMetricGroupConfig  `mapstructure:"session" yaml:"session" json:"session"`
	Size           SizeMetricGroupConfig     `mapstructure:"size" yaml:"size" json:"size"`
	Network        NetworkMetricGroupConfig  `mapstructure:"network" yaml:"network" json:"network"`
	Auth           AuthMetricsConfig         `mapstructure:"auth" yaml:"auth" json:"auth"`
}

type TimerConfig struct {
	TimerEnabled     bool `mapstructure:"timer_enabled" yaml:"timer_enabled" json:"timer_enabled"`
	HistogramEnabled bool `mapstructure:"histogram_enabled" yaml:"histogram_enabled" json:"histogram_enabled"`
}

type CounterConfig struct {
	OkEnabled    bool `mapstructure:"ok_enabled" yaml:"ok_enabled" json:"ok_enabled"`
	ErrorEnabled bool `mapstructure:"error_enabled" yaml:"error_enabled" json:"error_enabled"`
}

type RequestsMetricGroupConfig struct {
	Enabled      bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Counter      CounterConfig `mapstructure:"counter" yaml:"counter" json:"counter"`
	Timer        TimerConfig   `mapstructure:"timer" yaml:"timer" json:"timer"`
	FilteredTags []string      `mapstructure:"filtered_tags" yaml:"filtered_tags" json:"filtered_tags"`
}

type FdbMetricGroupConfig struct {
	Enabled      bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Counter      CounterConfig `mapstructure:"counter" yaml:"counter" json:"counter"`
	Timer        TimerConfig   `mapstructure:"timer" yaml:"timer" json:"timer"`
	FilteredTags []string      `mapstructure:"filtered_tags" yaml:"filtered_tags" json:"filtered_tags"`
}

type SearchMetricGroupConfig struct {
	Enabled      bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Counter      CounterConfig `mapstructure:"counter" yaml:"counter" json:"counter"`
	Timer        TimerConfig   `mapstructure:"timer" yaml:"timer" json:"timer"`
	FilteredTags []string      `mapstructure:"filtered_tags" yaml:"filtered_tags" json:"filtered_tags"`
}

type SessionMetricGroupConfig struct {
	Enabled      bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Counter      CounterConfig `mapstructure:"counter" yaml:"counter" json:"counter"`
	Timer        TimerConfig   `mapstructure:"timer" yaml:"timer" json:"timer"`
	FilteredTags []string      `mapstructure:"filtered_tags" yaml:"filtered_tags" json:"filtered_tags"`
}

type SizeMetricGroupConfig struct {
	Enabled      bool     `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Namespace    bool     `mapstructure:"namespace" yaml:"namespace" json:"namespace"`
	Project      bool     `mapstructure:"project" yaml:"project" json:"project"`
	Collection   bool     `mapstructure:"collection" yaml:"collection" json:"collection"`
	FilteredTags []string `mapstructure:"filtered_tags" yaml:"filtered_tags" json:"filtered_tags"`
}

type NetworkMetricGroupConfig struct {
	Enabled      bool     `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	FilteredTags []string `mapstructure:"filtered_tags" yaml:"filtered_tags" json:"filtered_tags"`
}

type AuthMetricsConfig struct {
	Enabled      bool     `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	FilteredTags []string `mapstructure:"filtered_tags" yaml:"filtered_tags" json:"filtered_tags"`
}

type ProfilingConfig struct {
	Enabled         bool `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	EnableCPU       bool `mapstructure:"enable_cpu" yaml:"enable_cpu" json:"enable_cpu"`
	EnableHeap      bool `mapstructure:"enable_heap" yaml:"enable_heap" json:"enable_heap"`
	EnableBlock     bool `mapstructure:"enable_block" yaml:"enable_block" json:"enable_block"`
	EnableMutex     bool `mapstructure:"enable_mutex" yaml:"enable_mutex" json:"enable_mutex"`
	EnableGoroutine bool `mapstructure:"enable_goroutine" yaml:"enable_goroutine" json:"enable_goroutine"`
}

type ManagementConfig struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
}

type ObservabilityConfig struct {
	Provider    string `mapstructure:"provider" yaml:"provider" json:"provider"`
	Enabled     bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	ApiKey      string `mapstructure:"api_key" yaml:"api_key" json:"api_key"`
	AppKey      string `mapstructure:"app_key" yaml:"app_key" json:"app_key"`
	ProviderUrl string `mapstructure:"provider_url" yaml:"provider_url" json:"provider_url"`
}

var (
	WriteUnitSize = 1024
	ReadUnitSize  = 4096
)

var DefaultConfig = Config{
	Log: log.LogConfig{
		Level:      "info",
		SampleRate: 0.01,
	},
	Server: ServerConfig{
		Type:         DatabaseServerType,
		Host:         "0.0.0.0",
		Port:         8081,
		RealtimePort: 8083,
		FDBHardDrop:  true,
	},
	Auth: AuthConfig{
		Enabled:          false,
		IssuerURL:        "https://tigrisdata-dev.us.auth0.com/",
		Audience:         "https://tigris-api",
		JWKSCacheTimeout: 5 * time.Minute,
		LogOnly:          true,
		AdminNamespaces:  []string{"tigris-admin"},
	},
	Cdc: CdcConfig{
		Enabled:        false,
		StreamInterval: 500 * time.Millisecond,
		StreamBatch:    100,
		StreamBuffer:   200,
	},
	Search: SearchConfig{
		Host:         "localhost",
		Port:         8108,
		ReadEnabled:  true,
		WriteEnabled: true,
	},
	Cache: CacheConfig{
		Host:    "0.0.0.0",
		Port:    6379,
		MaxScan: 500,
	},
	Tracing: TracingConfig{
		Enabled: false,
		Datadog: DatadogTracingConfig{
			Enabled:             false,
			SampleRate:          0.01,
			CodeHotspotsEnabled: true,
			EndpointsEnabled:    true,
		},
		Jaeger: JaegerTracingConfig{
			Enabled:    true,
			Url:        "http://tigris_jaeger:14268/api/traces",
			SampleRate: 0.01,
		},
	},
	Metrics: MetricsConfig{
		Enabled:        true,
		TimerQuantiles: []float64{0.5, 0.95, 0.99},
		Requests: RequestsMetricGroupConfig{
			Enabled: true,
			Counter: CounterConfig{
				OkEnabled:    true,
				ErrorEnabled: true,
			},
			Timer: TimerConfig{
				TimerEnabled:     true,
				HistogramEnabled: false,
			},
			FilteredTags: nil,
		},
		Fdb: FdbMetricGroupConfig{
			Enabled: true,
			Counter: CounterConfig{
				OkEnabled:    true,
				ErrorEnabled: true,
			},
			Timer: TimerConfig{
				TimerEnabled:     true,
				HistogramEnabled: false,
			},
			FilteredTags: nil,
		},
		Search: SearchMetricGroupConfig{
			Enabled: true,
			Counter: CounterConfig{
				OkEnabled:    true,
				ErrorEnabled: true,
			},
			Timer: TimerConfig{
				TimerEnabled:     true,
				HistogramEnabled: false,
			},
			FilteredTags: nil,
		},
		Session: SessionMetricGroupConfig{
			Enabled: true,
			Counter: CounterConfig{
				OkEnabled:    true,
				ErrorEnabled: true,
			},
			Timer: TimerConfig{
				TimerEnabled:     true,
				HistogramEnabled: false,
			},
			FilteredTags: nil,
		},
		Size: SizeMetricGroupConfig{
			Enabled:      true,
			Namespace:    true,
			Project:      true,
			Collection:   true,
			FilteredTags: nil,
		},
		Network: NetworkMetricGroupConfig{
			Enabled:      true,
			FilteredTags: nil,
		},
		Auth: AuthMetricsConfig{
			Enabled:      true,
			FilteredTags: nil,
		},
	},
	Profiling: ProfilingConfig{
		Enabled:    false,
		EnableCPU:  true,
		EnableHeap: true,
	},
	Quota: QuotaConfig{
		// Maximum limits single node can handle. Across namespaces.
		Node: LimitsConfig{
			Enabled: false,

			ReadUnits:  4000, // read requests per second per namespace
			WriteUnits: 1000, // write requests per second
		},
		Namespace: NamespaceLimitsConfig{
			// Per cluster limits for single namespace
			Default: LimitsConfig{
				Enabled: false,

				ReadUnits:  100, // read requests per second per namespace
				WriteUnits: 25,  // write requests per second
			},
			// Maximum per node quota for single namespace
			Node: LimitsConfig{
				ReadUnits:  100, // read requests per second per namespace
				WriteUnits: 25,  // write requests per second
			},
			RefreshInterval: 60 * time.Second,
			Regulator: QuotaRegulator{
				Increment:  5,
				Hysteresis: 10,
			},
		},
		Storage: StorageLimitsConfig{
			Enabled:         false,
			DataSizeLimit:   100 * 1024 * 1024,
			RefreshInterval: 60 * time.Second,
		},
	},
	Observability: ObservabilityConfig{
		Enabled:     false,
		Provider:    "datadog",
		ProviderUrl: "us3.datadoghq.com",
	},
	Management: ManagementConfig{
		Enabled: true,
	},
	Schema: SchemaConfig{
		AllowIncompatible: false,
	},
}

// SchemaConfig contains schema related settings.
type SchemaConfig struct {
	// AllowIncompatible when set to true enables incompatible schema changes support.
	//
	// Compatible schema change include:
	//	* adding new fields
	//  * adding defaults
	//  * extending max_length of the string fields
	//
	// Incompatible schema changes include:
	//	* deleting a field
	//  * changing field type
	//  * reducing max_length of the string fields
	//  * setting "required" property
	AllowIncompatible bool `mapstructure:"allow_incompatible" json:"allow_incompatible" yaml:"allow_incompatible"`
}

// FoundationDBConfig keeps FoundationDB configuration parameters.
type FoundationDBConfig struct {
	ClusterFile string `mapstructure:"cluster_file" json:"cluster_file" yaml:"cluster_file"`
}

type SearchConfig struct {
	Host         string `mapstructure:"host" json:"host" yaml:"host"`
	Port         int16  `mapstructure:"port" json:"port" yaml:"port"`
	AuthKey      string `mapstructure:"auth_key" json:"auth_key" yaml:"auth_key"`
	ReadEnabled  bool   `mapstructure:"read_enabled" yaml:"read_enabled" json:"read_enabled"`
	WriteEnabled bool   `mapstructure:"write_enabled" yaml:"write_enabled" json:"write_enabled"`
}

type CacheConfig struct {
	Host    string `mapstructure:"host" json:"host" yaml:"host"`
	Port    int16  `mapstructure:"port" json:"port" yaml:"port"`
	MaxScan int64  `mapstructure:"max_scan" json:"max_scan" yaml:"max_scan"`
}

type LimitsConfig struct {
	Enabled bool

	ReadUnits  int `mapstructure:"read_units" yaml:"read_units" json:"read_units"`
	WriteUnits int `mapstructure:"write_units" yaml:"write_units" json:"write_units"`
}

func (l *LimitsConfig) Limit(isWrite bool) int {
	if isWrite {
		return l.WriteUnits
	}

	return l.ReadUnits
}

type NamespaceLimitsConfig struct {
	Enabled    bool
	Default    LimitsConfig            // default per namespace limit
	Node       LimitsConfig            // max per node per namespace limit
	Namespaces map[string]LimitsConfig // individual namespaces configuration

	RefreshInterval time.Duration `mapstructure:"refresh_interval" yaml:"refresh_interval" json:"refresh_interval"`
	Regulator       QuotaRegulator
}

func (n *NamespaceLimitsConfig) NamespaceLimits(ns string) *LimitsConfig {
	cfg, ok := n.Namespaces[ns]
	if ok {
		return &cfg
	}
	return &n.Default
}

type NamespaceStorageLimitsConfig struct {
	Size int64
}

type StorageLimitsConfig struct {
	Enabled         bool
	DataSizeLimit   int64         `mapstructure:"data_size_limit" yaml:"data_size_limit" json:"data_size_limit"`
	RefreshInterval time.Duration `mapstructure:"refresh_interval" yaml:"refresh_interval" json:"refresh_interval"`

	// Per namespace limits
	Namespaces map[string]NamespaceStorageLimitsConfig
}

func (n *StorageLimitsConfig) NamespaceLimits(ns string) int64 {
	cfg, ok := n.Namespaces[ns]
	if ok {
		return cfg.Size
	}
	return n.DataSizeLimit
}

type QuotaRegulator struct {
	// This is a hysteresis band, deviation from ideal value in which regulation is no happening
	Hysteresis int `mapstructure:"hysteresis" yaml:"hysteresis" json:"hysteresis"`

	// when rate adjustment is needed increment current rate by ± this value
	// (this is percentage of maximum per node per namespace limit)
	// Set by config.DefaultConfig.Quota.Namespace.Node.(Read|Write)RateLimit.
	Increment int `mapstructure:"increment" yaml:"increment" json:"increment"`
}

type QuotaConfig struct {
	Node      LimitsConfig          // maximum rates per node. protects the node from overloading
	Namespace NamespaceLimitsConfig // user quota across all the nodes
	Storage   StorageLimitsConfig

	WriteUnitSize int
	ReadUnitSize  int
}

func (s *SearchConfig) IsReadEnabled() bool {
	return s.WriteEnabled && s.ReadEnabled
}
