package context

import (
	"context"

	"github.com/spf13/afero"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.persist-loki/pkg/lib/cmdflags"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.persist-loki/pkg/lib/logger"
)

// NewContext returns an initialized application context
func NewContext(ctx context.Context, fs *afero.Afero, flags cmdflags.Flags) Context {
	var level logger.Level
	if flags.Debug {
		level = logger.Debug
	} else {
		level = logger.Info
	}

	return Context{
		Ctx:    ctx,
		Fs:     fs,
		Logger: logger.New(level),
		Flags:  flags,
	}
}
