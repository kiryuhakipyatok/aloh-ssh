package main

import (
	"github.com/kiryuhakipyatok/aloh-ssh/internal/app"
)

func main() {
	// if err := godotenv.Load("../../.env"); err != nil {
	// 	panic(fmt.Errorf("failed to load env: %w", err))
	// }
	app.Run()
}
