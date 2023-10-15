package config

type Config struct {
	Endpoint       string
	ReportInterval int
	PollInterval   int
}

func New() (Config, error) {
	var c Config
	ParseFlag(&c)
	return c, nil
}
