package config

type Config struct {
    Port         string
    Host         string
    Directory    string
	Destination  string
    Username     string
    Password     string
    NgrokToken   string
	WithNgrok    bool
}

var Shared = &Config{}