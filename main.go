package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Grabs Network ID's; multiple versions representing how this is possible are located at the end of this script
// The Platform / Cloud Team will provide a config file listing environments for the CLI tool to build upon
// First, define the shape of the config file in a struct, followed by the actual function
type IDPConfig struct {
	VpcID    string `json:"vpc_id"`
	SubnetID string `json:"subnet_id"`
}

// TODO: In a production enterprise deployment, replace this local file read
// with an authenticated API call to HashiCorp Vault or AWS Secrets Manager.
// luke edit the above comment to reflect that this should be vaulted!
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

// Scan the local state directory to report open sandboxes
func listEnvironments() {
	stateDir := "./modules/idp-env/terraform.tfstate.d" //moved from foundation/

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
		// Native Workspaces use directories to isolate tracking metadata
		if file.IsDir() {
			envName := file.Name()

			// Skip the default system workspace room if it appears
			if envName == "default" {
				continue
			}

			// Fetch the metadata attributes of the folder
			info, err := file.Info()
			if err != nil {
				continue
			}

			// Calculate the age delta by subtracting its last modification time from right now
			age := time.Since(info.ModTime())

			// Format the age into a clean string (e.g., "Running for: 2h 15m")
			ageString := fmt.Sprintf("%dh %dm", int(age.Hours()), int(age.Minutes())%60)

			// THE SNARKY CURFEW MONITOR SYSTEM
			// If the environment was modified over 24 hours ago, throw a friendly nudge
			snarkyRemark := ""
			if age.Hours() > 24 {
				snarkyRemark = "Left the lights on, eh? Run < idp --name " + envName + " --destroy > to save our cloud budget!"
			} else {
				snarkyRemark = "Still a fresh sandbox. Happy Testing!"
			}

			// Print the aggregated metadata line
			fmt.Printf("Name:  %-20s [Last Modification: %s] %s\n", envName, ageString, snarkyRemark)
		}
	}

	fmt.Println("===========================================")
	fmt.Println("Reminder: Don't forget to clear up old environments when finished testing!")
}

// Terraform Handoff
func runTerraform(action string, envName string, vpcID string, subnetID string) error {
	fmt.Printf("Communicating with Terraform to execute: %s...\n\n", strings.ToUpper(action))
	// First, Ensure the workspace exists and select it
	// We run 'terraform workspace select <name>' or 'new' if it's missing.
	// Check if the workspace exists by attempting to select it first.
	selectCmd := exec.Command("terraform", "-chdir=./modules/idp-env", "workspace", "select", envName)
	// If selecting fails, it means the workspace doesn't exist yet! We create it.
	if err := selectCmd.Run(); err != nil {
		fmt.Printf("Creating fresh isolation workspace room: '%s'...\n", envName)
		newCmd := exec.Command("terraform", "-chdir=./modules/idp-env", "workspace", "new", envName)
		if err := newCmd.Run(); err != nil {
			return fmt.Errorf("Failed to create new state workspace: %v", err)
		}
	}

	// Run the lifecycle action inside that isolated workspace, Constructing the command-line execution string dynamically
	cmd := exec.Command("terraform", "-chdir=./modules/idp-env", action, "-auto-approve",
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

// Tests passed, here's where the magic happens
func main() {
	// Let them know stuff is happening
	fmt.Println("IDP-CLI: Initiating sandbox deployment...")

	//Get user input name flags
	args, err := parseFlags()
	if err != nil {
		fmt.Printf("Input error: %v\n", err)
		return
	}

	// Route to the List Processor if requested
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

	// Terraform Setup; Determine the core lifecycle action
	tfAction := "apply"
	if args.IsDestroy {
		tfAction = "destroy"
	}

	// Terraform Handoff (Flat Guard Clause)
	err = runTerraform(tfAction, args.EnvName, vpcID, subnetID)
	if err != nil {
		fmt.Printf("\nTerraform Failure: %v\n", err)
		return
	}

	// Happy Path
	if args.IsDestroy {
		fmt.Printf("Target ENV named %s has been torn down successfully!\n", args.EnvName)

		// DISK PURGE AUTOMATION: Clean up the local empty workspace folder from your hard drive
		// --destroy'd sandboxes were still showing in --list. This makes --list accurate after a successful --destroy.
		workspacePath := "./modules/idp-env/terraform.tfstate.d/" + args.EnvName
		fmt.Printf("🧹 Clearing empty workspace tracking data from disk: %s...\n", args.EnvName)
		_ = os.RemoveAll(workspacePath) // Quietly wipes out the empty folder!

	} else {
		fmt.Printf("Connected to Foundation VPC: %s\n", vpcID)
		fmt.Printf("Setup on Subnet: %s\n", subnetID)
		fmt.Printf("Test ENV: %s, Launched Successfully!\n", args.EnvName)
	}
}

// Extra Network ID - Getters
// For testing locally, if one had the master .hcl files, this is how Terraform would check
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
