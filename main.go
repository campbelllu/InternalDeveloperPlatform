package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Grabs Network ID's; multiple versions representing how this is possible are located at the end of this script
// The Platform / Cloud Team will provide a config file listing environments for the CLI tool to build upon
// First, define the shape of the config file in a struct, followed by the actual function
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

// Fetches CLI inputs for the EC2 tags / Env-names
// First we use a struct for inputs to keep the function clean as functionality via flags expands
type CLIArgs struct {
	EnvName   string
	IsDestroy bool
	IsList    bool
}

func parseFlags() (CLIArgs, error) {
	// Declare all available flags, default name to empty "", and add documentation text to all
	namePtr := flag.String("name", "", "The env name of your custom sandbox environment (Required, unless using --list)")
	destroyPtr := flag.Bool("destroy", false, "Set to true to tear down the specified environment")
	listPtr := flag.Bool("list", false, "Set to true to list all currently running IDP environments")

	// Tell Go to parse the incoming command-line flags
	flag.Parse()

	// Package the raw inputs into our structured container
	args := CLIArgs{
		EnvName:   strings.TrimSpace(*namePtr),
		IsDestroy: *destroyPtr,
		IsList:    *listPtr,
	}

	// If they want to list environments, they don't need to provide a name!
	if args.IsList {
		return args, nil //Listing doesn't require a name parameter
	}

	// If they aren't grabbing a list, then an environment name is strictly mandatory
	if args.EnvName == "" {
		return args, fmt.Errorf("Missing required parameter! You must provide an environment name via: --name <env-name> \n Or request a tracking pull via: --list")
	}

	return args, nil
}

// Luke test this once evn's exist
// Scan the local state directory to report open sandboxes
func listEnvironments() {
	stateDir := "./foundation/terraform.tfstate.d"

	// Open the directory and read its contents
	files, err := os.ReadDir(stateDir)

	if err != nil || len(files) == 0 {
		fmt.Println("\n===============================")
		fmt.Println("ZERO active environments found.")
		fmt.Println("===============================")
		return
	}

	fmt.Println("\n==========================================")
	fmt.Println("Active Sandboxes tracked by this platform:")
	fmt.Println("------------------------------------------")

	for _, file := range files {
		// Only look for files ending in .tfstate
		if !file.IsDir() && filepath.Ext(file.Name()) == ".tfstate" {
			// Strip the ".tfstate" extension off to print a clean name string
			cleanName := strings.TrimSuffix(file.Name(), ".tfstate")
			fmt.Printf(" %s\n", cleanName)
		}
	}
	fmt.Println("===========================================")
	fmt.Println("Reminder: Don't forget to clear up old environments when finished testing!")
}

// Luke this probably will need testing
// Terraform Handoff
func runTerraform(action string, envName string, vpcID string, subnetID string) error {
	fmt.Printf("Communicating with Terraform to execute: %s...\n\n", strings.ToUpper(action))

	// Construct the command-line execution string dynamically
	// Point Terraform to our reusable child module folder using -state mapping
	// cmd := exec.Command("terraform", action, "-auto-approve",
	// 	"-state=terraform.tfstate.d/"+envName+".tfstate", // Isolate each dev's tracking receipts!
	// 	"-var", "env_name="+envName,
	// 	"-var", "vpc_id="+vpcID,
	// 	"-var", "subnet_id="+subnetID,
	// 	"./modules/idp-env", // Target our reusable factory module folder
	// )
	cmd := exec.Command("terraform",
		"-chdir=./modules/idp-env", // 1. Global flag ALWAYS goes first to set the room
		action,                     // 2. The action (apply or destroy)
		"-auto-approve",            // 3. Automation flag
		// 4. We use "../../" because Terraform is now standing inside the module folder
		// and needs to reach back out to the root to save the dynamic state tracking slip!
		"-state=../../terraform.tfstate.d/"+envName+".tfstate",
		"-var", "env_name="+envName,
		"-var", "vpc_id="+vpcID,
		"-var", "subnet_id="+subnetID,
	)

	// THE SECRET SAUCE: Wire the background hidden screen directly to the active terminal
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run it
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Terraform execution failed: %v", err)
	}

	return nil
}

// Luke don't forget to test this all: create and destroy!
func main() {
	// Let them know stuff is happening
	fmt.Println("IDP-CLI: Initiating sandbox deployment...")

	//Get user's input name flags
	//Luke this was added and commented out
	args, err := parseFlags()
	// envName, err := parseFlags()
	if err != nil {
		fmt.Printf("Input error: %v\n", err)
		return
	}

	// luke this a and b were added with the above code
	// Step A: Route to the List Processor if requested
	if args.IsList {
		fmt.Println("Checking for active sandboxes...")
		listEnvironments()
		return
	}

	//Get Foundation ID's
	vpcID, subnetID, err := fetchFoundationIDs()
	if err != nil {
		fmt.Printf("Failed to secure network boundaries: %v\n", err)
		return
	}

	// Step B: Otherwise, proceed to standard Terraform Handoff ; Determine the core lifecycle action
	//luke there's an error because terraform logic below is commented out
	tfAction := "apply"
	if args.IsDestroy {
		tfAction = "destroy"
	}

	// 4. Execute the Terraform Handoff (Flat Guard Clause)
	err = runTerraform(tfAction, args.EnvName, vpcID, subnetID)
	if err != nil {
		fmt.Printf("\nTerraform Failure: %v\n", err)
		return
	}

	// Happy Path
	if args.IsDestroy {
		fmt.Printf("Target ENV named %s has been torn down successfully!\n", args.EnvName)
	} else {
		fmt.Printf("Connected to Foundation VPC: %s\n", vpcID)
		fmt.Printf("Setup on Subnet: %s\n", subnetID)
		fmt.Printf("Target ENV Name, Launched Successfully: %s\n", args.EnvName)
	}
}

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

// ##### del, but kept for now because i thought it was cool haha
// // 3. Glue the background command's screen directly to your active terminal
// cmd.Stdout = os.Stdout
// cmd.Stderr = os.Stderr
