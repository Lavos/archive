package main

import (
	"github.com/Lavos/archive"
)

func main () {
	s := archive.NewServer()
	s.Run()
}
