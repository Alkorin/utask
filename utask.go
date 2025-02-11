package utask

import (
	"context"
	"encoding/json"
	"fmt"

	"golang.org/x/sync/semaphore"

	"github.com/ovh/configstore"
)

var (
	// Version holds the tag of current µTask release
	Version string
	// Commit is the current git commit hash
	Commit string
	// App name (from configuration)
	App string
	// InstanceID identifies this running instance of µTask, as registered in DB
	InstanceID uint64

	// FInitializersFolder is the path to a folder containing
	// .so plugins for µTask initialization
	FInitializersFolder string
	// FPluginFolder is the path to a folder containing
	// .so plugins to be registered as step action executors
	FPluginFolder string
	// FTemplatesFolder is the path to a folder containing
	// .yaml templates for tasks
	FTemplatesFolder string
	// FRegion is the region in which this instance of µTask is running
	FRegion string
	// FPort is the port on which the http server listens
	FPort uint
	// FDebug is a flag to toggle debug log
	FDebug bool
	// FMaintenanceMode is a flag to prevent all write operations on the API,
	// except for admin actions (key rotation)
	FMaintenanceMode bool
)

// AppName returns the name of the application (from config)
func AppName() string { return App }

const (
	// DBName is the name of µTask DB, as registered on zesty
	DBName = "uservice_task"

	// MaxPageSize is the upper limit for the number of elements returned in a single page
	MaxPageSize = 10000
	// DefaultPageSize is the default number of elements returned in a single page
	DefaultPageSize = 1000
	// MinPageSize is the lower limit for the number of elements returned in a single page
	MinPageSize = 10

	// DefaultRetryMax is the default number of retries allowed for a task's execution
	DefaultRetryMax = 100

	// MaxTextSizeLong is the maximum number of characters accepted in a text-type field
	MaxTextSizeLong = 100000 // ~100 kB
	// MaxTextSize is the maximum number of characters accepted in a simple string field
	MaxTextSize = 1000 // ~1 kB
	// MinTextSize is the minimum number of characters accepted in any string-type field
	MinTextSize = 3

	// This is the key used in Values for a step to refer to itself
	This = "this"
)

// UtaskCfgSecretAlias is the key for the config item containing global configuration data
const UtaskCfgSecretAlias = "utask-cfg"

// Cfg holds global configuration data
type Cfg struct {
	ApplicationName         string                   `json:"application_name"`
	AdminUsernames          []string                 `json:"admin_usernames"`
	ResolverUsernames       []string                 `json:"resolver_usernames"`
	CompletedTaskExpiration string                   `json:"completed_task_expiration"`
	NotifyConfig            map[string]NotifyBackend `json:"notify_config"`
	NotifyActions           NotifyActions            `json:"notify_actions"`
	DatabaseConfig          *DatabaseConfig          `json:"database_config"`
	ConcealedSecrets        []string                 `json:"concealed_secrets"`
	ResourceLimits          map[string]uint          `json:"resource_limits"`
	MaxConcurrentExecutions uint                     `json:"max_concurrent_executions"`

	resourceSemaphores map[string]*semaphore.Weighted
	executionSemaphore *semaphore.Weighted
}

// NotifyBackend holds configuration for instantiating a notify client
type NotifyBackend struct {
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
}

// NotifyBackendTat holds configuration for instantiating a Tat notify client
type NotifyBackendTat struct {
	Username string `json:"username"`
	Password string `json:"password"`
	URL      string `json:"url"`
	Topic    string `json:"topic"`
}

// NotifyActions holds configuration of each actions
// By default all the actions are enabled /w any config name registered
type NotifyActions struct {
	TaskStateAction NotifyActionsParameters `json:"task_state_action,omitempty"`
}

// NotifyActionsParameters holds configuration needed to define each Notify actions
// If NotifyBackends is empty, the default is any
type NotifyActionsParameters struct {
	Disabled       bool     `json:"disabled"`
	NotifyBackends []string `json:"notify_backends"`
}

// DatabaseConfig holds configuration to fine-tune DB connection
type DatabaseConfig struct {
	MaxOpenConns    int `json:"max_open_conns"`
	MaxIdleConns    int `json:"max_idle_conns"`
	ConnMaxLifetime int `json:"conn_max_lifetime"`
}

func (c *Cfg) buildLimits() {
	c.resourceSemaphores = make(map[string]*semaphore.Weighted)

	for k, v := range c.ResourceLimits {
		c.resourceSemaphores[k] = semaphore.NewWeighted(int64(v))
	}

	if c.MaxConcurrentExecutions > 0 {
		c.executionSemaphore = semaphore.NewWeighted(int64(c.MaxConcurrentExecutions))
	}
}

// AcquireResource takes a semaphore slot for a named resource
// limiting the amount of concurrent actions runnable on said resource
func AcquireResource(name string) {
	if global == nil {
		return
	}
	s := global.resourceSemaphores[name]
	if s == nil {
		return
	}
	s.Acquire(context.Background(), 1)
}

// ReleaseResource frees up a semaphore slot for a named resource
func ReleaseResource(name string) {
	if global == nil {
		return
	}
	s := global.resourceSemaphores[name]
	if s == nil {
		return
	}
	s.Release(1)
}

// AcquireExecutionSlot takes a slot from a global semaphore
// putting a cap on the total amount of concurrent task executions
func AcquireExecutionSlot() {
	if global == nil {
		return
	}
	if global.executionSemaphore == nil {
		return
	}
	global.executionSemaphore.Acquire(context.Background(), 1)
}

// ReleaseExecutionSlot frees up a slot on the global execution semaphore
func ReleaseExecutionSlot() {
	if global == nil {
		return
	}
	if global.executionSemaphore == nil {
		return
	}
	global.executionSemaphore.Release(1)
}

var global *Cfg

// Config returns the global configuration data of this instance
// once lazy-loaded from configstore
func Config(store *configstore.Store) (*Cfg, error) {
	if global == nil {
		global = &Cfg{}

		cfgStr, err := configstore.Filter().Slice(UtaskCfgSecretAlias).Squash().Store(store).MustGetFirstItem().Value()
		if err != nil {
			return nil, fmt.Errorf("failed to get utask configuration from store: %s", err)
		}

		if err := json.Unmarshal([]byte(cfgStr), &global); err != nil {
			return nil, fmt.Errorf("failed to unmarshal utask configuration: %s", err)
		}

		App = global.ApplicationName

		global.buildLimits()
	}

	return global, nil
}
