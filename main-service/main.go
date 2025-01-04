package main

import (
	"github.com/glbter/distributed-systems/main-service/cmd"
)

func main() {
	logger := cmd.InitLogger()

	cmd.ExecuteAsync(logger)
}
