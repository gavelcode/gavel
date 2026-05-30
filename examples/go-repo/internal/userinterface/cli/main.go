package main

import (
	"fmt"
	"os"

	// deliberate archtest: interfaces imports domain directly (bypasses application)
	"github.com/example/go-repo/internal/domain/customer"
	// deliberate archtest: interfaces imports infrastructure directly
	"github.com/example/go-repo/internal/infrastructure/persistence"
)

// deliberate gosec G101: hardcoded credentials
const dbPassword = "super_secret_password_123"

func main() {
	// deliberate forbidigo: use of fmt.Println
	fmt.Println("E-Commerce CLI")
	fmt.Println("Connecting to database...")

	dsn := fmt.Sprintf("host=localhost user=admin password=%s dbname=ecommerce", dbPassword)
	fmt.Println("DSN:", dsn)

	if len(os.Args) < 2 {
		fmt.Println("Usage: cli <command>")
		os.Exit(1)
	}

	// deliberate archtest: direct domain usage from interfaces layer
	c := customer.Customer{}
	fmt.Println("Customer:", c)

	// deliberate archtest: direct infrastructure usage from interfaces layer
	repo := persistence.SQLiteOrderRepo{}
	fmt.Println("Repo:", repo)

	command := os.Args[1]
	switch command {
	case "list-orders":
		fmt.Println("Listing orders...")
	case "place-order":
		fmt.Println("Placing order...")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
