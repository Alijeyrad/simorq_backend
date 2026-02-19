package config

type Config struct {
	Database        DatabaseConfig       `mapstructure:"database"`
	ContextDatabase DatabaseConfig       `mapstructure:"context_database"`
	CasbinDatabase  DatabaseConfig       `mapstructure:"casbin_database"`
	Redis           RedisConfig          `mapstructure:"redis"`
	Server          ServerConfig         `mapstructure:"server"`
	Authentication  AuthenticationConfig `mapstructure:"authentication"`
	Authorization   AuthorizationConfig  `mapstructure:"authorization"`
	Email           EmailConfig          `mapstructure:"email"`
	SMS             SMSConfig            `mapstructure:"sms"`
	Password        PasswordConfig       `mapstructure:"password"`
	OTP             OTPConfig            `mapstructure:"otp"`
	Codes           CodesConfig          `mapstructure:"codes"`
	Observability   ObservabilityConfig  `mapstructure:"observability"`
	Logging         LoggingConfig        `mapstructure:"logging"`
	S3              S3Config             `mapstructure:"s3"`
	ZarinPal        ZarinPalConfig       `mapstructure:"zarinpal"`
	Nats            NatsConfig           `mapstructure:"nats"`
}

type NatsConfig struct {
	URL string `mapstructure:"url" yaml:"url"`
}

type DatabaseConfig struct {
	Host       string                  `mapstructure:"host"`
	Port       int                     `mapstructure:"port"`
	User       string                  `mapstructure:"user"`
	Password   string                  `mapstructure:"password"`
	DBName     string                  `mapstructure:"dbname"`
	SSLMode    string                  `mapstructure:"sslmode"`
	Pool       DatabasePoolConfig      `mapstructure:"pool"`
	Migrations DatabaseMigrationConfig `mapstructure:"migrations"`
	Logging    DatabaseLoggingConfig   `mapstructure:"logging"`
}

type DatabasePoolConfig struct {
	MaxOpenConns       int `mapstructure:"max_open_conns"`
	MaxIdleConns       int `mapstructure:"max_idle_conns"`
	ConnMaxLifetimeMin int `mapstructure:"conn_max_lifetime_minutes"`
}

type DatabaseMigrationConfig struct {
	AutoMigrate bool `mapstructure:"auto_migrate"`
	SafeMode    bool `mapstructure:"safe_mode"`
}

type DatabaseLoggingConfig struct {
	Enabled              bool `mapstructure:"enabled"`
	SlowQueryThresholdMs int  `mapstructure:"slow_query_threshold_ms"`
}

type RedisConfig struct {
	Addr                string `mapstructure:"addr"`
	DB                  int    `mapstructure:"db"`
	Username            string `mapstructure:"username"`
	Password            string `mapstructure:"password"`
	PoolSize            int    `mapstructure:"pool_size"`
	MinIdleConns        int    `mapstructure:"min_idle_conns"`
	DialTimeoutSeconds  int    `mapstructure:"dial_timeout_seconds"`
	ReadTimeoutSeconds  int    `mapstructure:"read_timeout_seconds"`
	WriteTimeoutSeconds int    `mapstructure:"write_timeout_seconds"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
}

type ServerConfig struct {
	Port           int           `mapstructure:"port"`
	TimeoutSeconds int           `mapstructure:"timeout_seconds"`
	Environment    string        `mapstructure:"environment"`
	Domain         string        `mapstructure:"domain"`
	Databases      []string      `mapstructure:"databases"`
	CORS           CORSConfig    `mapstructure:"cors"`
	Headers        HeadersConfig `mapstructure:"headers"`
}

type HeadersConfig struct {
	XSSProtection             string `mapstructure:"xss_protection"`
	ContentTypeNosniff        string `mapstructure:"content_type_nosniff"`
	XFrameOptions             string `mapstructure:"x_frame_options"`
	ReferrerPolicy            string `mapstructure:"referrer_policy"`
	CrossOriginEmbedderPolicy string `mapstructure:"cross_origin_embedder_policy"`
	CrossOriginOpenerPolicy   string `mapstructure:"cross_origin_opener_policy"`
	CrossOriginResourcePolicy string `mapstructure:"cross_origin_resource_policy"`
	OriginAgentCluster        string `mapstructure:"origin_agent_cluster"`
	XDNSPrefetchControl       string `mapstructure:"x_dns_prefetch_control"`
	XDownloadOptions          string `mapstructure:"x_download_options"`
	XPermittedCrossDomain     string `mapstructure:"x_permitted_cross_domain"`
}

type CORSConfig struct {
	Enabled          bool     `mapstructure:"enabled"`
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAgeSeconds    int      `mapstructure:"max_age_seconds"`
}

type AuthenticationConfig struct {
	DefaultPasswordLength int          `mapstructure:"default_password_length"`
	Paseto                PasetoConfig `mapstructure:"paseto"`
	SessionTTLMinutes     int          `mapstructure:"session_ttl_minutes"`
	OTPTTLMinutes         int          `mapstructure:"otp_ttl_minutes"`
	// EncryptionKey is a 32-byte hex string used for AES-256-GCM encryption
	// of sensitive fields such as national_id and IBAN.
	EncryptionKey string `mapstructure:"encryption_key"`
}

type PasetoConfig struct {
	Mode             string `mapstructure:"mode"`
	LocalKeyHex      string `mapstructure:"local_key_hex"`
	SecretKeyHex     string `mapstructure:"secret_key_hex"`
	PublicKeyHex     string `mapstructure:"public_key_hex"`
	Issuer           string `mapstructure:"issuer"`
	Audience         string `mapstructure:"audience"`
	AccessTTLMinutes int    `mapstructure:"access_ttl_minutes"`
	RefreshTTLDays   int    `mapstructure:"refresh_ttl_days"`
}

type AuthorizationConfig struct {
	CasbinModelPath    string `mapstructure:"casbin_model_path"`
	EnableAudit        bool   `mapstructure:"enable_audit"`
	SuperadminBypass   bool   `mapstructure:"superadmin_bypass"`
	PolicySyncEnabled  bool   `mapstructure:"policy_sync_enabled"`
	HealthCheckEnabled bool   `mapstructure:"health_check_enabled"`
}

type EmailConfig struct {
	Enabled bool       `mapstructure:"enabled"`
	From    string     `mapstructure:"from"`
	SMTP    SMTPConfig `mapstructure:"smtp"`
}

type SMTPConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	UseTLS         bool   `mapstructure:"use_tls"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
}

type SMSConfig struct {
	Enabled bool        `mapstructure:"enabled"`
	SMSIR   SMSIRConfig `mapstructure:"smsir"`
}

type SMSIRConfig struct {
	APIKey     string `mapstructure:"api_key"`
	SecretKey  string `mapstructure:"secret_key"`
	TemplateID string `mapstructure:"template_id"`
}

type PasswordConfig struct {
	Algorithm     string `mapstructure:"algorithm"`
	MemoryKiB     uint32 `mapstructure:"memory_kib"`
	Iterations    uint32 `mapstructure:"iterations"`
	Parallelism   uint8  `mapstructure:"parallelism"`
	SaltLength    uint32 `mapstructure:"salt_length"`
	KeyLength     uint32 `mapstructure:"key_length"`
	LowMemoryMode bool   `mapstructure:"low_memory_mode"`
}

type OTPConfig struct {
	DefaultLength int    `mapstructure:"default_length"`
	MinLength     int    `mapstructure:"min_length"`
	MaxLength     int    `mapstructure:"max_length"`
	HashAlgorithm string `mapstructure:"hash_algorithm"`
}

type CodesConfig struct {
	TokenByteLength int    `mapstructure:"token_byte_length"`
	URLSafeTokens   bool   `mapstructure:"url_safe_tokens"`
	Charset         string `mapstructure:"charset"`
}

type ObservabilityConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	ServiceName    string        `mapstructure:"service_name"`
	ServiceVersion string        `mapstructure:"service_version"`
	Tracing        TracingConfig `mapstructure:"tracing"`
	Metrics        MetricsConfig `mapstructure:"metrics"`
}

type TracingConfig struct {
	Enabled      bool    `mapstructure:"enabled"`
	OTLPEndpoint string  `mapstructure:"otlp_endpoint"`
	OTLPInsecure bool    `mapstructure:"otlp_insecure"`
	SamplingRate float64 `mapstructure:"sampling_rate"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

type LoggingConfig struct {
	Level  string       `mapstructure:"level"`  // debug, info, warn, error
	Format string       `mapstructure:"format"` // text, json
	Output OutputConfig `mapstructure:"output"`
}

type OutputConfig struct {
	Stdout bool          `mapstructure:"stdout"`
	File   FileLogConfig `mapstructure:"file"`
	Loki   LokiConfig    `mapstructure:"loki"`
}

type FileLogConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Path       string `mapstructure:"path"`        // e.g. "logs/app.log"
	MaxSizeMB  int    `mapstructure:"max_size_mb"` // rotate after N MB
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
	Compress   bool   `mapstructure:"compress"`
}

type LokiConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"` // e.g. "http://localhost:3100"
	Username string `mapstructure:"username"` // for Grafana Cloud basic auth
	Password string `mapstructure:"password"`
}

type ZarinPalConfig struct {
	CallbackURL string `mapstructure:"callback_url"`
	MerchantID  string `mapstructure:"merchant_id"`
	Sandbox     bool   `mapstructure:"sandbox"`
}

type S3Config struct {
	Endpoint        string `mapstructure:"endpoint"`
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	Bucket          string `mapstructure:"bucket"`
	PresignTTLSec   int    `mapstructure:"presign_ttl_sec"`
}

func (c *Config) Validate() error {

	return nil
}
