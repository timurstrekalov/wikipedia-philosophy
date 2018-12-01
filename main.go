package main

import (
	"github.com/timurstrekalov/wikipedia-philosophy/server"
)

func main() {
	s := server.NewServer()
	s.Run()
}
