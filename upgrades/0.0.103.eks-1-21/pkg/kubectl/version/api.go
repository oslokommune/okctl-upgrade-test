package version

import (
	"fmt"
	yamlv3 "gopkg.in/yaml.v3"
)

func ParseVersions(bytes []byte) (Versions, error) {
	versions := Versions{}

	err := yamlv3.Unmarshal(bytes, &versions)
	if err != nil {
		return Versions{}, fmt.Errorf("unmarshalling kubectl version info (%s): %w", string(bytes), err)
	}

	m := make(map[interface{}]interface{})
	_ = yamlv3.Unmarshal(bytes, &m)

	return versions, nil
}
