package main

import (
	"fmt"
)

type endpoint struct {
	host string
	port int
}

func (e *endpoint) String() string {
	return fmt.Sprintf("%s:%d", e.host, e.port)
}
