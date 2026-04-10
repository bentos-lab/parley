package core

import (
	"context"
	"testing"

	"github.com/bentos-lab/parley/core/contract"
	"github.com/bentos-lab/parley/core/debate"
	"github.com/stretchr/testify/require"
)

type stubLLM struct {
	jsonResponse        string
	jsonResponses       []string
	generateCalled      bool
	generateJSONCalled  bool
	generateJSONCalls   int
	lastReq             contract.LLMRequest
	lastSchema          *contract.LLMJSONSchema
	generateJSONReqs    []contract.LLMRequest
	generateJSONSchemas []*contract.LLMJSONSchema
}

func (s *stubLLM) Generate(ctx context.Context, req contract.LLMRequest) (string, error) {
	s.generateCalled = true
	s.lastReq = req
	return "unexpected", nil
}

func (s *stubLLM) GenerateJSON(ctx context.Context, req contract.LLMRequest, schema *contract.LLMJSONSchema) (string, error) {
	s.generateJSONCalled = true
	s.lastReq = req
	s.lastSchema = schema
	s.generateJSONCalls++
	s.generateJSONReqs = append(s.generateJSONReqs, req)
	s.generateJSONSchemas = append(s.generateJSONSchemas, schema)
	if len(s.jsonResponses) > 0 {
		response := s.jsonResponses[0]
		s.jsonResponses = s.jsonResponses[1:]
		return response, nil
	}
	return s.jsonResponse, nil
}

func (s *stubLLM) GenerateStream(ctx context.Context, req contract.LLMRequest) (<-chan contract.LLMStreamChunk, <-chan error) {
	chunkCh := make(chan contract.LLMStreamChunk)
	errCh := make(chan error, 1)
	close(chunkCh)
	close(errCh)
	return chunkCh, errCh
}

func TestGenerateRoundStoresStructuredFields(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	debateItem := &debate.Debate{
		Name:  "Sample Debate",
		Topic: "Testing topic",
		Agents: []debate.DebateAgent{
			{ID: "a1", Name: "Alex", Stance: "pro"},
		},
	}
	filename := "alpha.2026-03-31-22-34-54.json"
	require.NoError(t, debateItem.SaveAs(filename))
	llm := &stubLLM{
		jsonResponse: `{"weakness":"Weak point","new_point":"New point with evidence","rebuttal":"You said \"X\"","final_speak":"Final response text","summary":"Short summary."}`,
	}

	usecase := &CreateRoundUsecase{
		LLMResolver: contract.LLMResolverFunc(func(provider string, model string) (contract.LLM, error) {
			return llm, nil
		}),
		Defaults: LLMDefaults{
			Provider:    "openai",
			OpenAIModel: "model",
		},
	}
	roundOutput, err := usecase.Execute(context.Background(), CreateRoundInput{
		Filename: filename,
	})
	require.NoError(t, err)
	require.True(t, llm.generateJSONCalled)
	require.False(t, llm.generateCalled)
	require.NotNil(t, llm.lastSchema)

	round := roundOutput.Round
	require.Equal(t, "Final response text", round.Message)
	require.Equal(t, "Weak point", round.Weakness)
	require.Equal(t, "New point with evidence", round.NewPoint)
	require.Equal(t, "You said \"X\"", round.Rebuttal)
	require.Equal(t, "Short summary.", round.Summary)
}

func TestGenerateRoundAllowsEmptyRebuttalFields(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	debateItem := &debate.Debate{
		Name:  "Sample Debate",
		Topic: "Testing topic",
		Agents: []debate.DebateAgent{
			{ID: "a1", Name: "Alex", Stance: "pro"},
		},
	}
	filename := "alpha.2026-03-31-22-34-54.json"
	require.NoError(t, debateItem.SaveAs(filename))
	llm := &stubLLM{
		jsonResponse: `{"weakness":"","new_point":"Opening point","rebuttal":"","final_speak":"Opening response","summary":"Opening summary."}`,
	}

	usecase := &CreateRoundUsecase{
		LLMResolver: contract.LLMResolverFunc(func(provider string, model string) (contract.LLM, error) {
			return llm, nil
		}),
		Defaults: LLMDefaults{
			Provider:    "openai",
			OpenAIModel: "model",
		},
	}
	roundOutput, err := usecase.Execute(context.Background(), CreateRoundInput{
		Filename: filename,
	})
	require.NoError(t, err)
	round := roundOutput.Round
	require.Equal(t, "", round.Weakness)
	require.Equal(t, "Opening point", round.NewPoint)
	require.Equal(t, "", round.Rebuttal)
	require.Equal(t, "Opening response", round.Message)
	require.Equal(t, "Opening summary.", round.Summary)
}

// TestGenerateRoundUsesContextSelection verifies inbound context overrides debate-stored LLM selection.
// Parameters: t provides the test context.
func TestGenerateRoundUsesContextSelection(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	debateItem := &debate.Debate{
		Name:        "Sample Debate",
		Topic:       "Testing topic",
		LLMProvider: "anthropic",
		LLMModel:    "stored-model",
		Agents: []debate.DebateAgent{
			{ID: "a1", Name: "Alex", Stance: "pro"},
		},
	}
	filename := "alpha.2026-03-31-22-34-54.json"
	require.NoError(t, debateItem.SaveAs(filename))
	llm := &stubLLM{
		jsonResponse: `{"weakness":"Weak point","new_point":"New point with evidence","rebuttal":"You said \"X\"","final_speak":"Final response text","summary":"Short summary."}`,
	}
	resolvedProvider := ""
	resolvedModel := ""
	usecase := &CreateRoundUsecase{
		LLMResolver: contract.LLMResolverFunc(func(provider string, model string) (contract.LLM, error) {
			resolvedProvider = provider
			resolvedModel = model
			return llm, nil
		}),
		Defaults: LLMDefaults{
			Provider:       "openai",
			OpenAIModel:    "default-openai-model",
			AnthropicModel: "default-anthropic-model",
			GeminiModel:    "default-gemini-model",
		},
	}
	ctx := WithLLMSelection(context.Background(), "gemini", "ctx-model")
	_, err := usecase.Execute(ctx, CreateRoundInput{Filename: filename})
	require.NoError(t, err)
	require.Equal(t, "gemini", resolvedProvider)
	require.Equal(t, "ctx-model", resolvedModel)
}
