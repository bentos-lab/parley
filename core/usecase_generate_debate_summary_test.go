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
		jsonResponses: []string{
			`{"points":["Point A"]}`,
			`{"final_conclusion":"Conclusion A"}`,
		},
	}
	usecase := &GenerateDebateSummaryUsecase{
		LLMResolver: contract.LLMResolverFunc(func(provider string, model string) (contract.LLM, error) {
			return llm, nil
		}),
		LLMProvider: "test",
		Model:       "model",
	}
	output, err := usecase.Execute(context.Background(), GenerateDebateSummaryInput{Filename: filename})
	require.NoError(t, err)
	require.True(t, llm.generateJSONCalled)
	require.Equal(t, 2, llm.generateJSONCalls)
	require.Equal(t, "Conclusion A", output.Summary.Conclusion)
	require.Equal(t, [][]string{{"Point A"}}, output.Summary.Agents)

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
			Agents:     [][]string{{"Stored point"}},
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
			Agents:     [][]string{{"Old point"}},
			Conclusion: "Old conclusion",
		},
	}
	filename := "alpha.2026-03-31-22-34-54.json"
	require.NoError(t, debateItem.SaveAs(filename))

	llm := &stubLLM{
		jsonResponses: []string{
			`{"points":["New point"]}`,
			`{"final_conclusion":"New conclusion"}`,
		},
	}
	usecase := &GenerateDebateSummaryUsecase{
		LLMResolver: contract.LLMResolverFunc(func(provider string, model string) (contract.LLM, error) {
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
	require.Equal(t, 2, llm.generateJSONCalls)
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

// TestGenerateDebateSummarySplitsPerAgentAndConclusion verifies per-agent points use agent-only transcripts and conclusion uses full transcript.
// Parameters: t provides the test context.
func TestGenerateDebateSummarySplitsPerAgentAndConclusion(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	debateItem := &debate.Debate{
		Name:  "Alpha",
		Topic: "Topic A",
		Agents: []debate.DebateAgent{
			{ID: "agent-1", Name: "Alex", Stance: "pro"},
			{ID: "agent-2", Name: "Blake", Stance: "con"},
		},
		Rounds: []debate.DebateRound{
			{AgentID: "agent-1", Message: "Agent 1 message"},
			{AgentID: "", Message: "User message"},
			{AgentID: "agent-2", Message: "Agent 2 message"},
		},
	}
	filename := "alpha.2026-03-31-22-34-54.json"
	require.NoError(t, debateItem.SaveAs(filename))

	llm := &stubLLM{
		jsonResponses: []string{
			`{"points":["P1"]}`,
			`{"points":["P2"]}`,
			`{"final_conclusion":"Conclusion"}`,
		},
	}
	usecase := &GenerateDebateSummaryUsecase{
		LLMResolver: contract.LLMResolverFunc(func(provider string, model string) (contract.LLM, error) {
			return llm, nil
		}),
		LLMProvider: "test",
		Model:       "model",
	}

	output, err := usecase.Execute(context.Background(), GenerateDebateSummaryInput{Filename: filename})
	require.NoError(t, err)
	require.Equal(t, 3, llm.generateJSONCalls)
	require.Equal(t, [][]string{{"P1"}, {"P2"}}, output.Summary.Agents)
	require.Equal(t, "Conclusion", output.Summary.Conclusion)

	require.Len(t, llm.generateJSONReqs, 3)
	agent1Transcript := llm.generateJSONReqs[0].Messages[0].Content
	require.Contains(t, agent1Transcript, "Agent 1 message")
	require.NotContains(t, agent1Transcript, "Agent 2 message")
	require.NotContains(t, agent1Transcript, "User message")

	agent2Transcript := llm.generateJSONReqs[1].Messages[0].Content
	require.Contains(t, agent2Transcript, "Agent 2 message")
	require.NotContains(t, agent2Transcript, "Agent 1 message")
	require.NotContains(t, agent2Transcript, "User message")

	conclusionTranscript := llm.generateJSONReqs[2].Messages[0].Content
	require.Contains(t, conclusionTranscript, "Alex: Agent 1 message")
	require.Contains(t, conclusionTranscript, "User: User message")
	require.Contains(t, conclusionTranscript, "Blake: Agent 2 message")
}
