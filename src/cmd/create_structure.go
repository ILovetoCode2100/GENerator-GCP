package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/marklovelady/api-cli-generator/pkg/virtuoso"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newCreateStructureCmd() *cobra.Command {
	var (
		filename string
		dryRun   bool
	)
	
	cmd := &cobra.Command{
		Use:   "create-structure",
		Short: "Create complete test structure from file",
		Long: `Create a complete Virtuoso test structure from a YAML or JSON file.
		
This command reads a structure definition file and creates all the resources
in the correct order: project -> goals -> journeys -> checkpoints -> steps.

Use --dry-run to preview what would be created without actually creating anything.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Read the structure file
			data, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
			
			// Parse the structure
			var structure virtuoso.TestStructure
			
			// Try YAML first
			err = yaml.Unmarshal(data, &structure)
			if err != nil {
				// Try JSON
				err = json.Unmarshal(data, &structure)
				if err != nil {
					return fmt.Errorf("failed to parse file as YAML or JSON: %w", err)
				}
			}
			
			// Validate the structure
			if structure.Project.Name == "" {
				return fmt.Errorf("project name is required")
			}
			
			if len(structure.Goals) == 0 {
				return fmt.Errorf("at least one goal is required")
			}
			
			// Preview in dry run mode
			if dryRun {
				return previewStructure(&structure)
			}
			
			// Create the structure
			client := virtuoso.NewClient(cfg)
			resources, err := createStructure(client, &structure)
			if err != nil {
				return fmt.Errorf("failed to create structure: %w", err)
			}
			
			// Output results
			switch cfg.Output.DefaultFormat {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				encoder.Encode(resources)
				
			case "yaml":
				fmt.Printf("project_id: %d\n", resources.ProjectID)
				fmt.Printf("total_steps: %d\n", resources.TotalSteps)
				fmt.Printf("goals:\n")
				for _, g := range resources.Goals {
					fmt.Printf("  - id: %d\n", g.ID)
					fmt.Printf("    name: %s\n", g.Name)
					fmt.Printf("    journeys:\n")
					for _, j := range g.Journeys {
						fmt.Printf("      - id: %d\n", j.ID)
						fmt.Printf("        name: %s\n", j.Name)
						fmt.Printf("        checkpoints: %d\n", len(j.Checkpoints))
					}
				}
				
			case "ai":
				fmt.Printf("Successfully created test structure:\n\n")
				fmt.Printf("Project ID: %d\n", resources.ProjectID)
				fmt.Printf("Total resources created:\n")
				fmt.Printf("- Goals: %d\n", len(resources.Goals))
				journeyCount := 0
				checkpointCount := 0
				for _, g := range resources.Goals {
					journeyCount += len(g.Journeys)
					for _, j := range g.Journeys {
						checkpointCount += len(j.Checkpoints)
					}
				}
				fmt.Printf("- Journeys: %d\n", journeyCount)
				fmt.Printf("- Checkpoints: %d\n", checkpointCount)
				fmt.Printf("- Steps: %d\n", resources.TotalSteps)
				fmt.Printf("\nNext steps:\n")
				fmt.Printf("1. View in Virtuoso UI: https://app2.virtuoso.qa\n")
				fmt.Printf("2. Run the test journeys\n")
				fmt.Printf("3. Add more tests to the structure file\n")
				
			default: // human
				fmt.Printf("✅ Created test structure successfully!\n\n")
				fmt.Printf("📦 Project: %s (ID: %d)\n", structure.Project.Name, resources.ProjectID)
				fmt.Printf("\n📊 Summary:\n")
				fmt.Printf("   Goals created: %d\n", len(resources.Goals))
				
				journeyCount := 0
				checkpointCount := 0
				for _, g := range resources.Goals {
					journeyCount += len(g.Journeys)
					for _, j := range g.Journeys {
						checkpointCount += len(j.Checkpoints)
					}
				}
				fmt.Printf("   Journeys created: %d\n", journeyCount)
				fmt.Printf("   Checkpoints created: %d\n", checkpointCount)
				fmt.Printf("   Steps created: %d\n", resources.TotalSteps)
			}
			
			return nil
		},
	}
	
	cmd.Flags().StringVarP(&filename, "file", "f", "", "Structure definition file (YAML or JSON)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview what would be created without creating anything")
	cmd.MarkFlagRequired("file")
	
	return cmd
}

func previewStructure(structure *virtuoso.TestStructure) error {
	fmt.Printf("🔍 Preview mode - nothing will be created\n\n")
	fmt.Printf("Project: %s\n", structure.Project.Name)
	if structure.Project.Description != "" {
		fmt.Printf("  Description: %s\n", structure.Project.Description)
	}
	
	fmt.Printf("\nGoals (%d):\n", len(structure.Goals))
	for i, goal := range structure.Goals {
		fmt.Printf("  %d. %s (URL: %s)\n", i+1, goal.Name, goal.URL)
		fmt.Printf("     Journeys (%d):\n", len(goal.Journeys))
		
		for j, journey := range goal.Journeys {
			fmt.Printf("       %d. %s\n", j+1, journey.Name)
			fmt.Printf("          Checkpoints (%d):\n", len(journey.Checkpoints))
			
			for k, checkpoint := range journey.Checkpoints {
				fmt.Printf("            %d. %s (%d steps)\n", k+1, checkpoint.Name, len(checkpoint.Steps))
			}
		}
	}
	
	// Count totals
	journeyCount := 0
	checkpointCount := 0
	stepCount := 0
	for _, g := range structure.Goals {
		journeyCount += len(g.Journeys)
		for _, j := range g.Journeys {
			checkpointCount += len(j.Checkpoints)
			for _, c := range j.Checkpoints {
				stepCount += len(c.Steps)
			}
		}
	}
	
	fmt.Printf("\nTotals:\n")
	fmt.Printf("  Goals: %d\n", len(structure.Goals))
	fmt.Printf("  Journeys: %d\n", journeyCount)
	fmt.Printf("  Checkpoints: %d\n", checkpointCount)
	fmt.Printf("  Steps: %d\n", stepCount)
	
	return nil
}

func createStructure(client *virtuoso.Client, structure *virtuoso.TestStructure) (*virtuoso.CreatedResources, error) {
	resources := &virtuoso.CreatedResources{
		Goals: make([]virtuoso.CreatedGoal, 0),
	}
	
	// Create project
	fmt.Printf("Creating project: %s...\n", structure.Project.Name)
	project, err := client.CreateProject(structure.Project.Name, structure.Project.Description)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	resources.ProjectID = project.ID
	fmt.Printf("  ✓ Created project ID: %d\n", project.ID)
	
	// Create goals
	for _, goalDef := range structure.Goals {
		fmt.Printf("\nCreating goal: %s...\n", goalDef.Name)
		goal, err := client.CreateGoal(project.ID, goalDef.Name, goalDef.URL)
		if err != nil {
			return nil, fmt.Errorf("failed to create goal %s: %w", goalDef.Name, err)
		}
		fmt.Printf("  ✓ Created goal ID: %d\n", goal.ID)
		
		// Get snapshot ID
		snapshotID, err := client.GetGoalSnapshot(goal.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get snapshot for goal %d: %w", goal.ID, err)
		}
		
		createdGoal := virtuoso.CreatedGoal{
			ID:       goal.ID,
			Name:     goal.Name,
			Snapshot: snapshotID,
			Journeys: make([]virtuoso.CreatedJourney, 0),
		}
		
		// Convert snapshot ID to int for journey creation
		snapshotIDInt, err := strconv.Atoi(snapshotID)
		if err != nil {
			return nil, fmt.Errorf("invalid snapshot ID: %w", err)
		}
		
		// Create journeys
		for _, journeyDef := range goalDef.Journeys {
			fmt.Printf("  Creating journey: %s...\n", journeyDef.Name)
			journey, err := client.CreateJourney(goal.ID, snapshotIDInt, journeyDef.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to create journey %s: %w", journeyDef.Name, err)
			}
			fmt.Printf("    ✓ Created journey ID: %d\n", journey.ID)
			
			createdJourney := virtuoso.CreatedJourney{
				ID:          journey.ID,
				Name:        journey.Name,
				Checkpoints: make([]virtuoso.CreatedCheckpoint, 0),
			}
			
			// Create checkpoints
			for i, checkpointDef := range journeyDef.Checkpoints {
				fmt.Printf("    Creating checkpoint: %s...\n", checkpointDef.Name)
				checkpoint, err := client.CreateCheckpoint(goal.ID, snapshotIDInt, checkpointDef.Name)
				if err != nil {
					return nil, fmt.Errorf("failed to create checkpoint %s: %w", checkpointDef.Name, err)
				}
				fmt.Printf("      ✓ Created checkpoint ID: %d\n", checkpoint.ID)
				
				// Attach checkpoint to journey
				position := i + 2 // Positions start at 2
				err = client.AttachCheckpoint(journey.ID, checkpoint.ID, position)
				if err != nil {
					return nil, fmt.Errorf("failed to attach checkpoint %d to journey %d: %w", 
						checkpoint.ID, journey.ID, err)
				}
				
				createdCheckpoint := virtuoso.CreatedCheckpoint{
					ID:        checkpoint.ID,
					Name:      checkpoint.Title,
					StepCount: len(checkpointDef.Steps),
				}
				
				// Add steps
				for _, stepDef := range checkpointDef.Steps {
					switch stepDef.Type {
					case "navigate":
						_, err = client.AddNavigateStep(checkpoint.ID, stepDef.URL)
					case "click":
						_, err = client.AddClickStep(checkpoint.ID, stepDef.Selector)
					case "wait":
						timeout := stepDef.Timeout
						if timeout == 0 {
							timeout = 5000 // Default 5 seconds
						}
						_, err = client.AddWaitStep(checkpoint.ID, stepDef.Selector, timeout)
					default:
						return nil, fmt.Errorf("unknown step type: %s", stepDef.Type)
					}
					
					if err != nil {
						return nil, fmt.Errorf("failed to add %s step to checkpoint %d: %w", 
							stepDef.Type, checkpoint.ID, err)
					}
					resources.TotalSteps++
				}
				
				if len(checkpointDef.Steps) > 0 {
					fmt.Printf("      ✓ Added %d steps\n", len(checkpointDef.Steps))
				}
				
				createdJourney.Checkpoints = append(createdJourney.Checkpoints, createdCheckpoint)
			}
			
			createdGoal.Journeys = append(createdGoal.Journeys, createdJourney)
		}
		
		resources.Goals = append(resources.Goals, createdGoal)
	}
	
	return resources, nil
}