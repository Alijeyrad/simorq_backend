package authorize

import (
	"context"
	"log/slog"
	"sync/atomic"

	psqlwatcher "github.com/IguteChung/casbin-psql-watcher"
	casbin "github.com/casbin/casbin/v2"
	entadapter "github.com/casbin/ent-adapter"
)

// policyLoadHealthy tracks the health state of Casbin policy loading.
// When policy reload fails, this is set to false to trigger health check failures.
var policyLoadHealthy atomic.Bool

func init() {
	policyLoadHealthy.Store(true)
}

// IsPolicyHealthy returns true if the Casbin policy is in a healthy state.
// Returns false if the last policy reload attempt failed.
func IsPolicyHealthy() bool {
	return policyLoadHealthy.Load()
}

// CleanupFunc is a function that cleans up resources.
type CleanupFunc func(ctx context.Context)

// NewEnforcer creates a new Casbin DistributedEnforcer with PostgreSQL adapter and watcher.
// Returns the enforcer and a cleanup function that should be called on shutdown.
func NewEnforcer(modelPath string, dsn string) (*casbin.DistributedEnforcer, CleanupFunc, error) {
	a, err := entadapter.NewAdapter("postgres", dsn)
	if err != nil {
		return nil, nil, err
	}

	e, err := casbin.NewDistributedEnforcer(modelPath, a)
	if err != nil {
		return nil, nil, err
	}

	// Use PostgreSQL watcher for instant propagation
	w, err := psqlwatcher.NewWatcherWithConnString(context.Background(), dsn, psqlwatcher.Option{
		Channel: "casbin_policy_update",
	})
	if err != nil {
		return nil, nil, err
	}

	err = w.SetUpdateCallback(func(msg string) {
		slog.Debug("casbin policy update received", "message", msg)
		if err := e.LoadPolicy(); err != nil {
			slog.Error("failed to reload policy after watcher notification", "error", err)
			policyLoadHealthy.Store(false)
		} else {
			policyLoadHealthy.Store(true)
		}
	})
	if err != nil {
		return nil, nil, err
	}

	if err := e.SetWatcher(w); err != nil {
		return nil, nil, err
	}

	e.EnableAutoSave(true)
	e.EnableEnforce(true)

	// Cleanup function to close watcher connection
	cleanup := func(ctx context.Context) {
		if w != nil {
			slog.Info("closing casbin policy watcher")
			w.Close()
		}
		if e != nil {
			slog.Info("stopping casbin auto policy loading")
			e.StopAutoLoadPolicy()
		}
		slog.Info("casbin enforcer cleanup completed")
	}

	return e, cleanup, nil
}
