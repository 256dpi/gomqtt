package main

import (
	"flag"
)

type options struct {
	conf string
	help bool
}

func parseFlag() (*options, error) {
	opt := new(options)
	flag.StringVar(
		&opt.conf,
		"c",
		opt.conf,
		"Config file path",
	)
	flag.BoolVar(
		&opt.help,
		"h",
		false,
		"Show this help",
	)
	flag.Parse()
	return opt, nil
}
