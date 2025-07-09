package virtuoso

// TestStructure represents a complete test hierarchy
type TestStructure struct {
	Project ProjectDef `yaml:"project" json:"project"`
	Goals   []GoalDef  `yaml:"goals" json:"goals"`
}

// ProjectDef defines a project in the structure
type ProjectDef struct {
	ID          int    `yaml:"id,omitempty" json:"id,omitempty"`           // Optional existing project ID
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// GoalDef defines a goal in the structure
type GoalDef struct {
	Name     string       `yaml:"name" json:"name"`
	URL      string       `yaml:"url" json:"url"`
	Journeys []JourneyDef `yaml:"journeys" json:"journeys"`
}

// JourneyDef defines a journey in the structure
type JourneyDef struct {
	Name        string          `yaml:"name" json:"name"`
	Checkpoints []CheckpointDef `yaml:"checkpoints" json:"checkpoints"`
}

// CheckpointDef defines a checkpoint in the structure
type CheckpointDef struct {
	Name          string    `yaml:"name" json:"name"`
	NavigationURL string    `yaml:"navigation_url,omitempty" json:"navigation_url,omitempty"` // For first checkpoint only
	Steps         []StepDef `yaml:"steps" json:"steps"`
}

// StepDef defines a step in the structure
type StepDef struct {
	Type     string `yaml:"type" json:"type"`         // navigate, click, wait, fill
	URL      string `yaml:"url,omitempty" json:"url,omitempty"`
	Selector string `yaml:"selector,omitempty" json:"selector,omitempty"`
	Value    string `yaml:"value,omitempty" json:"value,omitempty"`     // For fill steps
	Timeout  int    `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// CreatedResources tracks all resources created during structure creation
type CreatedResources struct {
	ProjectID  int                   `json:"project_id"`
	Goals      []CreatedGoal         `json:"goals"`
	TotalSteps int                   `json:"total_steps"`
}

// CreatedGoal tracks a created goal and its resources
type CreatedGoal struct {
	ID        int               `json:"id"`
	Name      string            `json:"name"`
	Snapshot  string            `json:"snapshot_id"`
	Journeys  []CreatedJourney  `json:"journeys"`
}

// CreatedJourney tracks a created journey and its resources
type CreatedJourney struct {
	ID          int                  `json:"id"`
	Name        string               `json:"name"`
	Checkpoints []CreatedCheckpoint  `json:"checkpoints"`
}

// CreatedCheckpoint tracks a created checkpoint
type CreatedCheckpoint struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	StepCount int    `json:"step_count"`
}