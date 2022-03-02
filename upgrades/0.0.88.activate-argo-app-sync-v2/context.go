package main

import (
	"context"

	"github.com/oslokommune/okctl-upgrade/upgrades/0.0.88.activate-argo-app-sync-v2/pkg/logger"
	"github.com/spf13/afero"
)

type Context struct {
	Ctx    context.Context
	Fs     *afero.Afero
	logger logger.Logger
}

func newContext(ctx context.Context, fs *afero.Afero, flags cmdFlags) Context {
	var level logger.Level
	if flags.debug {
		level = logger.Debug
	} else {
		level = logger.Info
	}

	return Context{
		Ctx:    ctx,
		Fs:     fs,
		logger: logger.New(level),
	}
}
