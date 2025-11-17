package main

import (
	"flag"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	var password string
	flag.StringVar(&password, "password", "", "Password to hash")
	flag.Parse()

	if password == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -password <password>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s -password mySecurePassword123\n", os.Args[0])
		os.Exit(1)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error hashing password: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(hashed))
}
