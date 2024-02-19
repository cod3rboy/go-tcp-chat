package main

import "log"

func main() {
	app, err := createServer()
	if err != nil {
		log.Fatal(err)
	}

	app.Serve()
}
