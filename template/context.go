package main

import (
	"github.com/oslokommune/okctl-upgrade/template/pkg/logger"
)

type Context struct {
	logger logger.Logger
	force  bool
}

func newContext(flags cmdFlags) Context {
	var level logger.Level
	if flags.debug {
		level = logger.Debug
	} else {
		level = logger.Info
	}

	return Context{
		logger: logger.New(level),
		force:  flags.force,
	}
}
