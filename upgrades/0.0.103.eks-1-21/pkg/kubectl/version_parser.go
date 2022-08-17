package kubectl

import (
	"fmt"
	"strconv"
	"strings"

	yamlv3 "gopkg.in/yaml.v3"
)

func ParseVersion(bytes []byte) (Versions, error) {
	versions := Versions{}

	err := yamlv3.Unmarshal(bytes, &versions)
	if err != nil {
		return Versions{}, fmt.Errorf("unmarshalling kubectl version info (%s): %w", string(bytes), err)
	}

	m := make(map[interface{}]interface{})
	_ = yamlv3.Unmarshal(bytes, &m)

	return versions, nil
}

type Versions struct {
	ClientVersion Version `yaml:"clientVersion"`
	ServerVersion Version `yaml:"serverVersion"`
}

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
