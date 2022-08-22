package version

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Minor string `yaml:"minor"`
}

func (v Version) MinorAsInt() (int, error) {
	noPlus := strings.ReplaceAll(v.Minor, "+", "")

	num, err := strconv.Atoi(noPlus)
	if err != nil {
		return -1, fmt.Errorf("cannot parse '%s' to int: %w", noPlus, err)
	}

	return num, nil
}
