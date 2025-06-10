package main

import (
	"fmt"
	"os"

	verify "github.com/nubificus/secure-http-discovery-handler/internal/verify"
)

func main() {
	// get os.Args()[1] and os.Args()[2]
	if len(os.Args) < 3 {
		fmt.Println("Usage: verify-tool <deviceIP> <diceIP>")
		return
	}
	// args := os.Args[1:]
	deviceIP := os.Args[1]
	diceIP := os.Args[2]

	verified, err := verify.VerifyDevice(deviceIP, diceIP)
	if err != nil {
		fmt.Printf("Error verifying device: %v\n", err)
		return
	}
	if verified {
		fmt.Println("Device verified successfully.")
	} else {
		fmt.Println("Device verification failed.")
	}
}
