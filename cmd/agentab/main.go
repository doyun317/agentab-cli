package main

import (
	"context"
	"os"

	"github.com/agentab/agentab-cli/internal/app"
)

func main() {
	os.Exit(app.Run(context.Background(), os.Args[1:]))
}
