package somecomponent

import "github.com/oslokommune/okctl-upgrade/template/pkg/logger"

type SomeComponent struct {
	logger logger.Logger
	force  bool
}

func (c SomeComponent) Run() error {
	c.logger.Info("Upgrading SomeComponent")

	c.logger.Debug("SomeComponent is on version 0.5. Updating to 0.6")

	if c.force {
		// Actually do the upgrade
		c.logger.Debug("Woho!")
	}

	c.logger.Info("Upgrading SomeComponent complete!")

	return nil
}

func New(logger logger.Logger, force bool) SomeComponent {
	return SomeComponent{
		logger: logger,
		force:  force,
	}
}
