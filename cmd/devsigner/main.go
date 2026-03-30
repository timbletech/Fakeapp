package main

import (
	"device_only/internal/crypto"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load() // Loads .env from local path if devsigner is run in root

	signCmd := flag.NewFlagSet("sign", flag.ExitOnError)
	challengeOpt := signCmd.String("challenge", "", "Base64 challenge to sign")

	keygenCmd := flag.NewFlagSet("keygen", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("Expected 'sign' or 'keygen' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "keygen":
		keygenCmd.Parse(os.Args[2:])
		priv, pub, err := crypto.GenerateKeyPair()
		if err != nil {
			log.Fatalf("Keygen failed: %v", err)
		}
		fmt.Printf("DEV_PRIVATE_KEY=\n%s\n\nPUBLIC_KEY=\n%s\n", priv, pub)

	case "sign":
		signCmd.Parse(os.Args[2:])
		if *challengeOpt == "" {
			fmt.Println("Please provide a --challenge")
			os.Exit(1)
		}

		privKey := os.Getenv("DEV_PRIVATE_KEY")
		if privKey == "" {
			log.Fatal("DEV_PRIVATE_KEY environment variable is not set")
		}

		sig, err := crypto.SignChallenge(privKey, *challengeOpt)
		if err != nil {
			log.Fatalf("Failed to sign challenge: %v", err)
		}

		fmt.Println(sig)

	default:
		fmt.Println("Expected 'sign' or 'keygen' subcommands")
		os.Exit(1)
	}
}
