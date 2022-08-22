package version

type Versions struct {
	ClientVersion Version `yaml:"clientVersion"`
	ServerVersion Version `yaml:"serverVersion"`
}
