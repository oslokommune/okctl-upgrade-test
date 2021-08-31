package main

import (
	"github.com/oslokommune/okctl-upgrade/template/pkg/somecomponent"
)

func upgrade(context Context) error {
	c := somecomponent.New(context.logger, context.force)

	err := c.Run()
	if err != nil {
		return err
	}

	return nil
}
