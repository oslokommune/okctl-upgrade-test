// Package context helps share the most common aspects
package context

import (
	"context"

	"github.com/spf13/afero"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.persist-loki/pkg/lib/cmdflags"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.94.persist-loki/pkg/lib/logger"
)

// Context defines the application context
type Context struct {
	Ctx    context.Context
	Fs     *afero.Afero
	Logger logger.Logger
	Flags  cmdflags.Flags
}
