package main

import (
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.89.rotate-argocd-ssh-key/pkg/lib/cmdflags"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.89.rotate-argocd-ssh-key/pkg/lib/logger"
)

type Context struct {
	logger logger.Logger
}

func newContext(flags cmdflags.Flags) Context {
	var level logger.Level
	if flags.Debug {
		level = logger.Debug
	} else {
		level = logger.Info
	}

	return Context{
		logger: logger.New(level),
	}
}
