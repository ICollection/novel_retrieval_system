package main

import (
	"fmt"
	"service/internal/server"
)

func main() {
	for handler := range server.Engine.Handlers {
		fmt.Println(handler)
	}
}
