package somecomponent

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/oslokommune/okctl-upgrade/template/pkg/lib/cmdflags"
	"github.com/oslokommune/okctl-upgrade/template/pkg/lib/commonerrors"
	"github.com/oslokommune/okctl-upgrade/template/pkg/lib/logger"
)

// SomeComponent is a sample okctl component
type SomeComponent struct {
	flags   cmdflags.Flags
	log     logger.Logger
	dryRun  bool
	confirm bool
}

// Upgrade upgrades the component
func (c SomeComponent) Upgrade() error {
	c.log.Info("Upgrading SomeComponent")

	c.log.Debug("SomeComponent is on version 0.5. Updating to 0.6")

	if !c.dryRun && !c.confirm {
		c.log.Info("This will delete all logs.")

		answer, err := c.askUser("Do you want to continue?")
		if err != nil {
			return fmt.Errorf("prompting user: %w", err)
		}

		if !answer {
			return commonerrors.ErrUserAborted
		}
	}

	if c.dryRun {
		c.log.Info("Simulating some stuff")
	} else {
		c.log.Info("Doing some stuff")
	}

	c.log.Info("Upgrading SomeComponent done!")

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

func New(logger logger.Logger, flags cmdflags.Flags) SomeComponent {
	return SomeComponent{
		log:     logger,
		dryRun:  flags.DryRun,
		confirm: flags.Confirm,
	}
}
