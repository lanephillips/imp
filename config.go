package main

import "code.google.com/p/gcfg"

type Config struct {
	Server struct {
		Host string
		Port string
		Certificate string
		Key string
	}
	Database struct {
		Database string
		User string
		Password string
	}
	Mail struct {
		Name string
		Address string
		User string
		Host string
		Password string
	}
}

func LoadConfigInto(config *Config, file string) (err error) {
	err = gcfg.ReadFileInto(config, file)
	return
}