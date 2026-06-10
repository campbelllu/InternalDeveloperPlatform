package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Grabs Network ID's; multiple versions representing how this is possible are located at the end of this script
// The Platform / Cloud Team will provide a config file listing environments for the CLI tool to build upon
// First, define the shape of the config file, followed by the actual function
type IDPConfig struct {
	VpcID    string `json:"vpc_id"`
	SubnetID string `json:"subnet_id"`
}

func fetchFoundationIDs() (string, string, error) {
	fmt.Println("Fetching network boundaries from local configuration file...")

	// Open the config file
	configFile, err := os.ReadFile(".idp-config.json")

	if err != nil {
		return "", "", fmt.Errorf("Missing configuration file! Ensure platform setup is complete: %v", err)
	}

	// Parse the JSON text into a real Go structure
	var config IDPConfig
	err = json.Unmarshal(configFile, &config)

	if err != nil {
		return "", "", fmt.Errorf("Corrupted configuration file format: %v", err)
	}

	return config.VpcID, config.SubnetID, nil
}

// ### functions left to be made
// Fetches CLI inputs for the EC2 tags / Env-names

// Runs Terraform to create desired EC2 / env
// ### functions left to be made

func main() {
	// Let them know stuff is happening
	fmt.Println("IDP-CLI: Initiating sandbox deployment...")

	//Get Foundation ID's
	vpcID, subnetID, err := fetchFoundationIDs()

	// Oops?
	if err != nil {
		fmt.Printf("Failed to secure network boundaries: %v\n", err)
		return
	}

	// Happy Path
	fmt.Printf("Connected to Foundation VPC: %s\n", vpcID)
	fmt.Printf("ENV setup on subnet: %s\n", subnetID)

	// // 1. Simulated inputs (These will come from CLI flags later)
	// envName := "luke-dev-env"
	// vpcID := "vpc-04b6acb3c384dc94d"
	// subnetID := "subnet-091289cfe3586f82e"

	// // 2. Format the exact terminal command you would type manually
	// cmd := exec.Command("terraform", "apply", "-auto-approve",
	// 	"-var", "env_name="+envName,
	// 	"-var", "vpc_id="+vpcID,
	// 	"-var", "subnet_id="+subnetID,
	// )

	// // 3. Glue the background command's screen directly to your active terminal
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr

	// // 4. Pull the trigger
	// err := cmd.Run()
	// if err != nil {
	// 	fmt.Printf("Deployment failed: %v\n", err)
	// 	return
	// }

	// fmt.Println("Sandbox successfully deployed and secured via identity endpoints!")
}

// Test function upon first making file. Go works! :)
// func main() {
// 	fmt.Println("IDP-CLI Engine Initialized Successfully!")
// 	fmt.Println("-------------------------------------------")
// 	fmt.Println("Next step: Building the Terraform orchestrator.")
// }

// Extra Network ID - Getters
// For testing locally, if one had the master .hcl files, this is how Terraform would check the S3 state bucket
// and grab the necessary ID's
func adminFetchFoundationIDs() (string, string, error) {
	fmt.Println("Querying foundational infrastructure outputs...")

	// Fetch the live VPC ID from the foundation layer outputs
	vpcCmd := exec.Command("terraform", "-chdir=./foundation", "output", "-raw", "vpc_id")
	vpcOutput, err := vpcCmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to read VPC output: %v, details: %s", err, string(vpcOutput))
	}

	// Fetch the live Subnet ID from the foundation layer outputs
	subnetCmd := exec.Command("terraform", "-chdir=./foundation", "output", "-raw", "public_subnet_id")
	subnetOutput, err := subnetCmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("failed to read Subnet output: %v, details: %s", err, string(subnetOutput))
	}

	// Clean up the trailing whitespace/newlines that Terraform outputs leave behind
	vpcID := strings.TrimSpace(string(vpcOutput))
	subnetID := strings.TrimSpace(string(subnetOutput))

	return vpcID, subnetID, nil
}

// If the ID's were hardcoded into the CLI tool, not the best option as the dev / test env may be relocated randomly
// which would throw off devs without further CLI code updates
func hardCodedFetchFoundationIDs() (string, string, error) {
	// placeholders(ph), if we weren't fetching these
	phVpcID := "vpc-04b6acb3c384dc94d"
	phSubnetID := "subnet-091289cfe3586f82e"

	return phVpcID, phSubnetID, nil
}
