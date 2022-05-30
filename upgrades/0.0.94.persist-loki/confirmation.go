package main

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
)

const confirmationMessage = `Warning!

This upgrade will restart Loki. If persistent logging has not been set up, this will destroy existing logs.

Learn how to set up persistence here: https://oslokommune.slack.com/archives/CV9EGL9UG/p1629718366030700

Proceed with the upgrade?`

func requestConfirmation() (bool, error) {
	answer := false
	prompt := &survey.Confirm{Message: confirmationMessage}

	err := survey.AskOne(prompt, &answer)
	if err != nil {
		return false, fmt.Errorf("querying user: %w", err)
	}

	return answer, nil
}
