package config

type Config struct {
	Endpoint string
}

func New() (Config, error) {
	var c Config
	ParseFlag(&c)
	return c, nil
}
