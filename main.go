package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("IDP-CLI: Initiating sandbox deployment...")

	// 1. Simulated inputs (These will come from CLI flags later)
	envName := "luke-dev-env"
	vpcID := "vpc-04b6acb3c384dc94d"
	subnetID := "subnet-091289cfe3586f82e"

	// 2. Format the exact terminal command you would type manually
	cmd := exec.Command("terraform", "apply", "-auto-approve",
		"-var", "env_name="+envName,
		"-var", "vpc_id="+vpcID,
		"-var", "subnet_id="+subnetID,
	)

	// 3. Glue the background command's screen directly to your active terminal
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 4. Pull the trigger
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Deployment failed: %v\n", err)
		return
	}

	fmt.Println("Sandbox successfully deployed and secured via identity endpoints!")
}

// Test function upon first making file. Go works! :)
// func main() {
// 	fmt.Println("IDP-CLI Engine Initialized Successfully!")
// 	fmt.Println("-------------------------------------------")
// 	fmt.Println("Next step: Building the Terraform orchestrator.")
// }
