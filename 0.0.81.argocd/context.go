package main

import (
	"github.com/oslokommune/okctl-upgrade/0.0.81.argocd/pkg/logger"
)

type Context struct {
	log logger.Logger
}

func newContext(flags cmdFlags) Context {
	var level logger.Level
	if flags.debug {
		level = logger.Debug
	} else {
		level = logger.Info
	}

	return Context{
		log: logger.New(level),
	}
}
