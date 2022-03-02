package main

import (
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.87.argocd-v2/pkg/lib/cmdflags"
	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.87.argocd-v2/pkg/lib/logger"
)

type Context struct {
	log logger.Logger
}

func newContext(flags cmdflags.Flags) Context {
	var level logger.Level
	if flags.Debug {
		level = logger.Debug
	} else {
		level = logger.Info
	}

	return Context{
		log: logger.New(level),
	}
}
