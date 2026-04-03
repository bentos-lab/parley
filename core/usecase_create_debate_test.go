package core

import (
	"context"
	"testing"

	"github.com/bentos-lab/parley/core/debate"
	"github.com/stretchr/testify/require"
)

// TestCreateAssignsMissingAgentIDsAllMissing verifies IDs are generated when all are empty.
// Parameters: t is the test handle.
// Returns: nothing.
func TestCreateAssignsMissingAgentIDsAllMissing(t *testing.T) {
	t.Parallel()
	usecase := &CreateDebateUsecase{DefaultTTSProvider: "native"}
	input := CreateDebateInput{
		Name:  "Sample Debate",
		Topic: "Sample Topic",
		Agents: []debate.DebateAgent{
			{Name: "Agent One"},
			{Name: "Agent Two"},
		},
	}
	output, err := usecase.Execute(context.Background(), input)
	require.NoError(t, err)
	require.Len(t, output.Debate.Agents, 2)
	require.Equal(t, "agent-1", output.Debate.Agents[0].ID)
	require.Equal(t, "agent-2", output.Debate.Agents[1].ID)
}

// TestCreateAssignsMissingAgentIDsMixed verifies existing IDs are preserved and blanks are filled.
// Parameters: t is the test handle.
// Returns: nothing.
func TestCreateAssignsMissingAgentIDsMixed(t *testing.T) {
	t.Parallel()
	usecase := &CreateDebateUsecase{DefaultTTSProvider: "native"}
	input := CreateDebateInput{
		Name:  "Mixed IDs",
		Topic: "Sample Topic",
		Agents: []debate.DebateAgent{
			{ID: "custom-1", Name: "Agent One"},
			{Name: "Agent Two"},
			{Name: "Agent Three"},
		},
	}
	output, err := usecase.Execute(context.Background(), input)
	require.NoError(t, err)
	require.Len(t, output.Debate.Agents, 3)
	require.Equal(t, "custom-1", output.Debate.Agents[0].ID)
	require.Equal(t, "agent-1", output.Debate.Agents[1].ID)
	require.Equal(t, "agent-2", output.Debate.Agents[2].ID)
}

// TestCreateAssignsMissingAgentIDsCollision verifies generated IDs skip existing collisions.
// Parameters: t is the test handle.
// Returns: nothing.
func TestCreateAssignsMissingAgentIDsCollision(t *testing.T) {
	t.Parallel()
	usecase := &CreateDebateUsecase{DefaultTTSProvider: "native"}
	input := CreateDebateInput{
		Name:  "Collision IDs",
		Topic: "Sample Topic",
		Agents: []debate.DebateAgent{
			{ID: "agent-1", Name: "Agent One"},
			{Name: "Agent Two"},
		},
	}
	output, err := usecase.Execute(context.Background(), input)
	require.NoError(t, err)
	require.Len(t, output.Debate.Agents, 2)
	require.Equal(t, "agent-1", output.Debate.Agents[0].ID)
	require.Equal(t, "agent-2", output.Debate.Agents[1].ID)
}

// TestCreateDebateRequiresName verifies missing names fail creation.
func TestCreateDebateRequiresName(t *testing.T) {
	t.Parallel()
	usecase := &CreateDebateUsecase{DefaultTTSProvider: "native"}
	input := CreateDebateInput{
		Topic: "Sample Topic",
		Agents: []debate.DebateAgent{
			{Name: "Agent One"},
		},
	}
	_, err := usecase.Execute(context.Background(), input)
	require.Error(t, err)
}

// TestCreateDebateRequiresAgents verifies missing agents fail creation.
func TestCreateDebateRequiresAgents(t *testing.T) {
	t.Parallel()
	usecase := &CreateDebateUsecase{DefaultTTSProvider: "native"}
	input := CreateDebateInput{
		Name:  "Sample Debate",
		Topic: "Sample Topic",
	}
	_, err := usecase.Execute(context.Background(), input)
	require.Error(t, err)
}

func TestCreateDebateUsecaseDefaultsToConfigProvider(t *testing.T) {
	t.Parallel()
	usecase := &CreateDebateUsecase{DefaultTTSProvider: "native"}
	input := CreateDebateInput{
		Name:  "Default Provider",
		Topic: "Sample Topic",
		Agents: []debate.DebateAgent{
			{Name: "Agent One"},
		},
	}
	output, err := usecase.Execute(context.Background(), input)
	require.NoError(t, err)
	require.Equal(t, "native", output.Debate.TTSProvider)
}
