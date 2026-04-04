package core

import (
	"context"
	"testing"

	"github.com/bentos-lab/parley/core/contract"
	"github.com/bentos-lab/parley/core/debate"
	"github.com/stretchr/testify/require"
)

// TestGenerateDebateSummaryCreatesSummary verifies summaries are generated and persisted.
// Parameters: t provides the test context.
func TestGenerateDebateSummaryCreatesSummary(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	debateItem := &debate.Debate{
		Name:  "Alpha",
		Topic: "Topic A",
		Agents: []debate.DebateAgent{
			{ID: "agent-1", Name: "Alex", Stance: "pro"},
		},
		Rounds: []debate.DebateRound{
			{AgentID: "agent-1", Message: "Round 1"},
		},
	}
	filename := "alpha.2026-03-31-22-34-54.json"
	require.NoError(t, debateItem.SaveAs(filename))

	llm := &stubLLM{
		jsonResponse: `{"agents":{"agent-1":["Point A"]},"conclusion":"Conclusion A"}`,
	}
	usecase := &GenerateDebateSummaryUsecase{
		LLMResolver: contract.ResolverFunc[contract.LLM](func(name string) (contract.LLM, error) {
			return llm, nil
		}),
		LLMProvider: "test",
		Model:       "model",
	}
	output, err := usecase.Execute(context.Background(), GenerateDebateSummaryInput{Filename: filename})
	require.NoError(t, err)
	require.True(t, llm.generateJSONCalled)
	require.Equal(t, "Conclusion A", output.Summary.Conclusion)
	require.Equal(t, []string{"Point A"}, output.Summary.Agents["agent-1"])

	loaded, err := debate.LoadDebate(filename)
	require.NoError(t, err)
	require.NotNil(t, loaded.Summary)
	require.Equal(t, "Conclusion A", loaded.Summary.Conclusion)
}

// TestGenerateDebateSummaryReusesStoredSummary verifies summaries are reused when force is false.
// Parameters: t provides the test context.
func TestGenerateDebateSummaryReusesStoredSummary(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	debateItem := &debate.Debate{
		Name:  "Alpha",
		Topic: "Topic A",
		Agents: []debate.DebateAgent{
			{ID: "agent-1", Name: "Alex", Stance: "pro"},
		},
		Rounds: []debate.DebateRound{
			{AgentID: "agent-1", Message: "Round 1"},
		},
		Summary: &debate.DebateSummaryDetail{
			Agents: map[string][]string{
				"agent-1": {"Stored point"},
			},
			Conclusion: "Stored conclusion",
		},
	}
	filename := "alpha.2026-03-31-22-34-54.json"
	require.NoError(t, debateItem.SaveAs(filename))

	usecase := &GenerateDebateSummaryUsecase{}
	output, err := usecase.Execute(context.Background(), GenerateDebateSummaryInput{Filename: filename})
	require.NoError(t, err)
	require.Equal(t, "Stored conclusion", output.Summary.Conclusion)
}

// TestGenerateDebateSummaryForceNewRegenerates verifies force new regenerates the summary.
// Parameters: t provides the test context.
func TestGenerateDebateSummaryForceNewRegenerates(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	debateItem := &debate.Debate{
		Name:  "Alpha",
		Topic: "Topic A",
		Agents: []debate.DebateAgent{
			{ID: "agent-1", Name: "Alex", Stance: "pro"},
		},
		Rounds: []debate.DebateRound{
			{AgentID: "agent-1", Message: "Round 1"},
		},
		Summary: &debate.DebateSummaryDetail{
			Agents: map[string][]string{
				"agent-1": {"Old point"},
			},
			Conclusion: "Old conclusion",
		},
	}
	filename := "alpha.2026-03-31-22-34-54.json"
	require.NoError(t, debateItem.SaveAs(filename))

	llm := &stubLLM{
		jsonResponse: `{"agents":{"agent-1":["New point"]},"conclusion":"New conclusion"}`,
	}
	usecase := &GenerateDebateSummaryUsecase{
		LLMResolver: contract.ResolverFunc[contract.LLM](func(name string) (contract.LLM, error) {
			return llm, nil
		}),
		LLMProvider: "test",
		Model:       "model",
	}
	output, err := usecase.Execute(context.Background(), GenerateDebateSummaryInput{
		Filename: filename,
		ForceNew: true,
	})
	require.NoError(t, err)
	require.True(t, llm.generateJSONCalled)
	require.Equal(t, "New conclusion", output.Summary.Conclusion)
}

// TestGenerateDebateSummaryNoRoundsReturnsError verifies empty debates return an error.
// Parameters: t provides the test context.
func TestGenerateDebateSummaryNoRoundsReturnsError(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	debateItem := &debate.Debate{
		Name:  "Alpha",
		Topic: "Topic A",
	}
	filename := "alpha.2026-03-31-22-34-54.json"
	require.NoError(t, debateItem.SaveAs(filename))

	usecase := &GenerateDebateSummaryUsecase{}
	_, err := usecase.Execute(context.Background(), GenerateDebateSummaryInput{Filename: filename})
	require.Error(t, err)
}
