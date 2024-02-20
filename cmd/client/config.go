package main

import (
	"flag"

	"github.com/vharitonsky/iniflags"
)

// server connection config
var (
	host = flag.String("host", "localhost", "Server hostname")
	port = flag.Int("port", 4000, "Server port number")
)

func init() {
	iniflags.Parse()
}
