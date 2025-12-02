package main

import (
	"context"
	"fmt"

	"github.com/projuktisheba/ajfses/backend/api"
)

// startup is called at application startup
func main() {
	// Run backend server in main goroutine so the process stays alive
	ctx := context.Background()
	// Start backend server
	if err := api.RunServer(ctx); err != nil {
		fmt.Printf("Failed to start backend server: %v\n", err)
	}
}
