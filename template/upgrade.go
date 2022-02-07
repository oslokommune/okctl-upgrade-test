package main

import (
	"github.com/oslokommune/okctl-upgrade/template/pkg/lib/cmdflags"
	"github.com/oslokommune/okctl-upgrade/template/pkg/somecomponent"
)

func upgrade(context Context, flags cmdflags.Flags) error {
	c := somecomponent.New(context.logger, flags)

	err := c.Upgrade()
	if err != nil {
		return err
	}

	return nil
}
