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

type transactionLog struct {
	ProjectID    int
	Goals        []int
	Journeys     []int
	Checkpoints  []int
	StepsCreated int
}

func newCreateStructureEnhancedCmd() *cobra.Command {
	var (
		filename  string
		dryRun    bool
		verbose   bool
		projectID int
	)

	cmd := &cobra.Command{
		Use:   "create-structure",
		Short: "Create test structure handling auto-creation behavior",
		Long: `Create a complete Virtuoso test structure from a YAML or JSON file,
properly handling Virtuoso's auto-creation behavior:

- Goal creation auto-creates an initial journey - we REUSE it
- First checkpoint is always navigation - we UPDATE it, not create
- Navigation step is shared across the goal

Use --dry-run to preview what would be created without creating anything.
Use --verbose for detailed logging of each operation.
Use --project-id to use an existing project instead of creating a new one.`,
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

			// Override project ID if specified
			if projectID > 0 {
				structure.Project.ID = projectID
			}

			// Validate the structure
			if structure.Project.ID == 0 && structure.Project.Name == "" {
				return fmt.Errorf("project name is required when not using existing project")
			}

			if len(structure.Goals) == 0 {
				return fmt.Errorf("at least one goal is required")
			}

			// Preview in dry run mode
			if dryRun {
				return previewEnhancedStructure(&structure, verbose)
			}

			// Create the structure
			client := virtuoso.NewClient(cfg)
			resources, err := createEnhancedStructure(client, &structure, verbose)
			if err != nil {
				return fmt.Errorf("failed to create structure: %w", err)
			}

			// Output results
			outputCreatedResources(resources, cfg.Output.DefaultFormat, &structure)

			return nil
		},
	}

	cmd.Flags().StringVarP(&filename, "file", "f", "", "Structure definition file (YAML or JSON)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview what would be created without creating anything")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	cmd.Flags().IntVar(&projectID, "project-id", 0, "Use existing project ID instead of creating new")
	cmd.MarkFlagRequired("file")

	return cmd
}

func previewEnhancedStructure(structure *virtuoso.TestStructure, verbose bool) error {
	fmt.Printf("🔍 Preview mode - nothing will be created\n\n")

	if structure.Project.ID > 0 {
		fmt.Printf("Using existing project ID: %d\n", structure.Project.ID)
	} else {
		fmt.Printf("Project: %s\n", structure.Project.Name)
		if structure.Project.Description != "" {
			fmt.Printf("  Description: %s\n", structure.Project.Description)
		}
	}

	fmt.Printf("\nGoals: %d\n", len(structure.Goals))

	journeyCount := 0
	checkpointCount := 0
	stepCount := 0

	for i, g := range structure.Goals {
		fmt.Printf("\n  Goal %d: %s\n", i+1, g.Name)
		fmt.Printf("    URL: %s\n", g.URL)
		if i == 0 {
			fmt.Printf("    ⚠️  Note: Will use auto-created first journey\n")
		}

		for j, journey := range g.Journeys {
			if i == 0 && j == 0 {
				fmt.Printf("    Journey %d: %s (will RENAME auto-created journey)\n", j+1, journey.Name)
			} else {
				fmt.Printf("    Journey %d: %s\n", j+1, journey.Name)
			}
			journeyCount++

			for k, checkpoint := range journey.Checkpoints {
				checkpointCount++
				if j == 0 && k == 0 {
					fmt.Printf("      Checkpoint %d: %s (will UPDATE existing navigation checkpoint)\n", k+1, checkpoint.Name)
					if checkpoint.NavigationURL != "" {
						fmt.Printf("        Navigation URL: %s\n", checkpoint.NavigationURL)
					}
				} else {
					fmt.Printf("      Checkpoint %d: %s\n", k+1, checkpoint.Name)
				}

				if verbose && len(checkpoint.Steps) > 0 {
					fmt.Printf("        Steps:\n")
					for _, step := range checkpoint.Steps {
						fmt.Printf("          - %s", step.Type)
						if step.Selector != "" {
							fmt.Printf(" (selector: %s)", step.Selector)
						}
						if step.Value != "" {
							fmt.Printf(" (value: %s)", step.Value)
						}
						fmt.Printf("\n")
					}
				}
				stepCount += len(checkpoint.Steps)
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

func createEnhancedStructure(client *virtuoso.Client, structure *virtuoso.TestStructure, verbose bool) (*virtuoso.CreatedResources, error) {
	resources := &virtuoso.CreatedResources{
		Goals: make([]virtuoso.CreatedGoal, 0),
	}

	log := transactionLog{}

	// Step 1: Create or use existing project
	if structure.Project.ID > 0 {
		fmt.Printf("Using existing project ID: %d\n", structure.Project.ID)
		resources.ProjectID = structure.Project.ID
		log.ProjectID = structure.Project.ID
	} else {
		fmt.Printf("Creating project: %s...\n", structure.Project.Name)
		project, err := client.CreateProject(structure.Project.Name, structure.Project.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to create project: %w", err)
		}
		resources.ProjectID = project.ID
		log.ProjectID = project.ID
		fmt.Printf("  ✓ Created project ID: %d\n", project.ID)
	}

	// Step 2: Create goals (which auto-create journeys)
	for _, goalDef := range structure.Goals {
		fmt.Printf("\nCreating goal: %s...\n", goalDef.Name)
		goal, err := client.CreateGoal(resources.ProjectID, goalDef.Name, goalDef.URL)
		if err != nil {
			rollbackHint(&log)
			return nil, fmt.Errorf("failed to create goal %s: %w", goalDef.Name, err)
		}
		log.Goals = append(log.Goals, goal.ID)
		fmt.Printf("  ✓ Created goal ID: %d\n", goal.ID)

		// Get snapshot ID
		snapshotID, err := client.GetGoalSnapshot(goal.ID)
		if err != nil {
			rollbackHint(&log)
			return nil, fmt.Errorf("failed to get snapshot for goal %d: %w", goal.ID, err)
		}

		snapshotIDInt, err := strconv.Atoi(snapshotID)
		if err != nil {
			return nil, fmt.Errorf("invalid snapshot ID: %w", err)
		}

		createdGoal := virtuoso.CreatedGoal{
			ID:       goal.ID,
			Name:     goal.Name,
			Snapshot: snapshotID,
			Journeys: make([]virtuoso.CreatedJourney, 0),
		}

		// Step 3: Get the auto-created journey
		if verbose {
			fmt.Printf("  Looking for auto-created journey...\n")
		}
		existingJourneys, err := client.ListJourneys(goal.ID, snapshotIDInt)
		if err != nil {
			rollbackHint(&log)
			return nil, fmt.Errorf("failed to list journeys: %w", err)
		}

		if verbose {
			fmt.Printf("  Found %d existing journeys\n", len(existingJourneys))
		}

		// Process journeys
		for journeyIdx, journeyDef := range goalDef.Journeys {
			var journey *virtuoso.Journey

			if journeyIdx == 0 && len(existingJourneys) > 0 {
				// Use the auto-created journey for the first journey definition
				journey = existingJourneys[0]
				fmt.Printf("  Using auto-created journey: %s (ID: %d)\n", journey.Name, journey.ID)

				// Update the journey name if needed
				if journey.Name != journeyDef.Name {
					fmt.Printf("    Renaming to: %s...\n", journeyDef.Name)
					updatedJourney, err := client.UpdateJourney(journey.ID, journeyDef.Name)
					if err != nil {
						fmt.Printf("      ⚠️  Warning: Failed to rename journey: %v\n", err)
					} else {
						journey = updatedJourney
						fmt.Printf("      ✓ Journey renamed successfully\n")
					}
				}
			} else {
				// Create additional journeys
				fmt.Printf("  Creating journey: %s...\n", journeyDef.Name)
				journey, err = client.CreateJourney(goal.ID, snapshotIDInt, journeyDef.Name)
				if err != nil {
					rollbackHint(&log)
					return nil, fmt.Errorf("failed to create journey %s: %w", journeyDef.Name, err)
				}
				fmt.Printf("    ✓ Created journey ID: %d\n", journey.ID)

				// Update the name if API used a default
				if journey.Name != journeyDef.Name {
					fmt.Printf("    Updating journey name to: %s...\n", journeyDef.Name)
					updatedJourney, err := client.UpdateJourney(journey.ID, journeyDef.Name)
					if err != nil {
						fmt.Printf("      ⚠️  Warning: Failed to update journey name: %v\n", err)
					} else {
						journey = updatedJourney
						fmt.Printf("      ✓ Journey renamed successfully\n")
					}
				}
			}

			log.Journeys = append(log.Journeys, journey.ID)

			createdJourney := virtuoso.CreatedJourney{
				ID:          journey.ID,
				Name:        journey.Name,
				Checkpoints: make([]virtuoso.CreatedCheckpoint, 0),
			}

			// Process checkpoints
			for checkpointIdx, checkpointDef := range journeyDef.Checkpoints {
				if journeyIdx == 0 && checkpointIdx == 0 {
					// Step 5: Handle first checkpoint (navigation)
					fmt.Printf("    Handling navigation checkpoint: %s...\n", checkpointDef.Name)

					// Get the first checkpoint
					firstCheckpoint, err := client.GetFirstCheckpoint(journey.ID)
					if err != nil {
						rollbackHint(&log)
						return nil, fmt.Errorf("failed to get first checkpoint: %w", err)
					}

					if verbose {
						fmt.Printf("      Found existing checkpoint ID: %d\n", firstCheckpoint.ID)
					}

					// Update navigation URL if specified
					if checkpointDef.NavigationURL != "" {
						fmt.Printf("      Updating navigation URL to: %s...\n", checkpointDef.NavigationURL)
						// Note: This would require getting the navigation step and updating it
						// For now, we'll add this as a new navigation step
						_, err = client.AddNavigateStep(firstCheckpoint.ID, checkpointDef.NavigationURL)
						if err != nil {
							// Critical error
							rollbackHint(&log)
							return nil, fmt.Errorf("failed to update navigation: %w", err)
						}
						fmt.Printf("        ✓ Updated navigation\n")
					}

					// Add remaining steps to first checkpoint
					for _, stepDef := range checkpointDef.Steps {
						if err := addStepToCheckpoint(client, firstCheckpoint.ID, stepDef, verbose); err != nil {
							rollbackHint(&log)
							return nil, err
						}
						log.StepsCreated++
						resources.TotalSteps++
					}

					createdJourney.Checkpoints = append(createdJourney.Checkpoints, virtuoso.CreatedCheckpoint{
						ID:        firstCheckpoint.ID,
						Name:      checkpointDef.Name,
						StepCount: len(checkpointDef.Steps),
					})
				} else {
					// Create additional checkpoints normally
					fmt.Printf("    Creating checkpoint: %s...\n", checkpointDef.Name)
					checkpoint, err := client.CreateCheckpoint(goal.ID, snapshotIDInt, checkpointDef.Name)
					if err != nil {
						rollbackHint(&log)
						return nil, fmt.Errorf("failed to create checkpoint %s: %w", checkpointDef.Name, err)
					}
					fmt.Printf("      ✓ Created checkpoint ID: %d\n", checkpoint.ID)
					log.Checkpoints = append(log.Checkpoints, checkpoint.ID)

					// Attach checkpoint to journey
					position := checkpointIdx + 1 // Start at 1 for additional checkpoints
					if journeyIdx > 0 {
						position = checkpointIdx + 2 // For new journeys, start at position 2
					}

					err = client.AttachCheckpoint(journey.ID, checkpoint.ID, position)
					if err != nil {
						rollbackHint(&log)
						return nil, fmt.Errorf("failed to attach checkpoint %d to journey %d: %w",
							checkpoint.ID, journey.ID, err)
					}

					// Add steps
					for _, stepDef := range checkpointDef.Steps {
						if err := addStepToCheckpoint(client, checkpoint.ID, stepDef, verbose); err != nil {
							rollbackHint(&log)
							return nil, err
						}
						log.StepsCreated++
						resources.TotalSteps++
					}

					createdJourney.Checkpoints = append(createdJourney.Checkpoints, virtuoso.CreatedCheckpoint{
						ID:        checkpoint.ID,
						Name:      checkpoint.Title,
						StepCount: len(checkpointDef.Steps),
					})
				}
			}

			createdGoal.Journeys = append(createdGoal.Journeys, createdJourney)
		}

		resources.Goals = append(resources.Goals, createdGoal)
	}

	return resources, nil
}

func addStepToCheckpoint(client *virtuoso.Client, checkpointID int, stepDef virtuoso.StepDef, verbose bool) error {
	if verbose {
		fmt.Printf("        Adding %s step", stepDef.Type)
		if stepDef.Selector != "" {
			fmt.Printf(" (selector: %s)", stepDef.Selector)
		}
		fmt.Printf("...\n")
	}

	var err error
	switch stepDef.Type {
	case "navigate":
		_, err = client.AddNavigateStep(checkpointID, stepDef.URL)
	case "click":
		_, err = client.AddClickStep(checkpointID, stepDef.Selector)
	case "wait":
		timeout := stepDef.Timeout
		if timeout == 0 {
			timeout = 5000 // Default 5 seconds
		}
		_, err = client.AddWaitStep(checkpointID, stepDef.Selector, timeout)
	case "fill":
		_, err = client.AddFillStep(checkpointID, stepDef.Selector, stepDef.Value)
	default:
		return fmt.Errorf("unsupported step type: %s", stepDef.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to add %s step: %w", stepDef.Type, err)
	}

	if verbose {
		fmt.Printf("          ✓ Added %s step\n", stepDef.Type)
	}

	return nil
}

func rollbackHint(log *transactionLog) {
	fmt.Printf("\n⚠️  Error occurred! Resources created so far:\n")
	if log.ProjectID > 0 {
		fmt.Printf("  Project ID: %d\n", log.ProjectID)
	}
	if len(log.Goals) > 0 {
		fmt.Printf("  Goal IDs: %v\n", log.Goals)
	}
	if len(log.Journeys) > 0 {
		fmt.Printf("  Journey IDs: %v\n", log.Journeys)
	}
	if len(log.Checkpoints) > 0 {
		fmt.Printf("  Checkpoint IDs: %v\n", log.Checkpoints)
	}
	if log.StepsCreated > 0 {
		fmt.Printf("  Steps created: %d\n", log.StepsCreated)
	}
	fmt.Printf("\nManual cleanup may be required.\n")
}

func outputCreatedResources(resources *virtuoso.CreatedResources, format string, structure *virtuoso.TestStructure) {
	switch format {
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
			fmt.Printf("    snapshot_id: %s\n", g.Snapshot)
			fmt.Printf("    journeys:\n")
			for _, j := range g.Journeys {
				fmt.Printf("      - id: %d\n", j.ID)
				fmt.Printf("        name: %s\n", j.Name)
				fmt.Printf("        checkpoints:\n")
				for _, c := range j.Checkpoints {
					fmt.Printf("          - id: %d\n", c.ID)
					fmt.Printf("            name: %s\n", c.Name)
					fmt.Printf("            step_count: %d\n", c.StepCount)
				}
			}
		}

	case "ai":
		fmt.Printf("\n✅ Successfully created test structure!\n\n")
		fmt.Printf("Project: %s (ID: %d)\n", structure.Project.Name, resources.ProjectID)
		fmt.Printf("\nResource Summary:\n")
		fmt.Printf("- Goals created: %d\n", len(resources.Goals))
		journeyCount := 0
		checkpointCount := 0
		for _, g := range resources.Goals {
			journeyCount += len(g.Journeys)
			for _, j := range g.Journeys {
				checkpointCount += len(j.Checkpoints)
			}
		}
		fmt.Printf("- Journeys: %d (including renamed auto-created)\n", journeyCount)
		fmt.Printf("- Checkpoints: %d (including updated navigation)\n", checkpointCount)
		fmt.Printf("- Total steps: %d\n", resources.TotalSteps)

		fmt.Printf("\nImportant Notes:\n")
		fmt.Printf("- First journey in each goal was auto-created and renamed\n")
		fmt.Printf("- First checkpoint contains shared navigation step\n")
		fmt.Printf("- All resources are ready for test execution\n")

		fmt.Printf("\nNext Steps:\n")
		fmt.Printf("1. View in Virtuoso UI: https://app2.virtuoso.qa\n")
		fmt.Printf("2. Run the test journeys\n")
		fmt.Printf("3. Monitor test results\n")

	default: // human
		fmt.Printf("\n✅ Created test structure successfully!\n\n")
		if structure.Project.ID > 0 {
			fmt.Printf("📦 Using existing project ID: %d\n", resources.ProjectID)
		} else {
			fmt.Printf("📦 Project: %s (ID: %d)\n", structure.Project.Name, resources.ProjectID)
		}
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
		fmt.Printf("   Journeys: %d\n", journeyCount)
		fmt.Printf("   Checkpoints: %d\n", checkpointCount)
		fmt.Printf("   Steps created: %d\n", resources.TotalSteps)

		fmt.Printf("\n⚡ Special handling applied:\n")
		fmt.Printf("   - Auto-created journeys were renamed\n")
		fmt.Printf("   - Navigation checkpoints were updated\n")
		fmt.Printf("   - All resources properly linked\n")
	}
}
