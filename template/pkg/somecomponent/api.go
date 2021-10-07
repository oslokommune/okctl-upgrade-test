package somecomponent

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/oslokommune/okctl-upgrade/template/pkg/commonerrors"
	"github.com/oslokommune/okctl-upgrade/template/pkg/logger"
)

// SomeComponent is a sample okctl component
type SomeComponent struct {
	logger  logger.Logger
	dryRun  bool
	confirm bool
}

// Upgrade upgrades the component
func (c SomeComponent) Upgrade() error {
	c.logger.Info("Upgrading SomeComponent")

	c.logger.Debug("SomeComponent is on version 0.5. Updating to 0.6")

	if !c.dryRun && !c.confirm {
		c.logger.Info("This will delete all logs.")

		answer, err := c.askUser("Do you want to proceed?")
		if err != nil {
			return fmt.Errorf("prompting user: %w", err)
		}

		if !answer {
			return commonerrors.ErrUserAborted
		}
	}

	if c.dryRun {
		c.logger.Info("Simulating some stuff")
	} else {
		c.logger.Info("Doing some stuff")
	}

	c.logger.Info("Upgrading SomeComponent done!")

	return nil
}

func (c SomeComponent) askUser(question string) (bool, error) {
	answer := false
	prompt := &survey.Confirm{
		Message: question,
	}

	err := survey.AskOne(prompt, &answer)
	if err != nil {
		return false, err
	}

	return answer, nil
}

type Opts struct {
	DryRun  bool
	Confirm bool
}

func New(logger logger.Logger, opts Opts) SomeComponent {
	return SomeComponent{
		logger:  logger,
		dryRun:  opts.DryRun,
		confirm: opts.Confirm,
	}
}
