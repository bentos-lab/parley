package core

import (
	"context"
	"testing"

	"github.com/bentos-lab/parley/core/contract"
	"github.com/stretchr/testify/require"
)

type stubNameLLM struct {
	generated string
	lastReq   contract.LLMRequest
}

func (s *stubNameLLM) Generate(ctx context.Context, req contract.LLMRequest) (string, error) {
	s.lastReq = req
	return s.generated, nil
}

func (s *stubNameLLM) GenerateJSON(ctx context.Context, req contract.LLMRequest, schema *contract.LLMJSONSchema) (string, error) {
	return "", nil
}

func (s *stubNameLLM) GenerateStream(ctx context.Context, req contract.LLMRequest) (<-chan contract.LLMStreamChunk, <-chan error) {
	chunkCh := make(chan contract.LLMStreamChunk)
	errCh := make(chan error, 1)
	close(chunkCh)
	close(errCh)
	return chunkCh, errCh
}

func TestGenerateDebateNameUsecase(t *testing.T) {
	t.Parallel()
	llm := &stubNameLLM{generated: "Future of AI"}
	usecase := &GenerateDebateNameUsecase{
		LLMResolver: contract.ResolverFunc[contract.LLM](func(name string) (contract.LLM, error) {
			return llm, nil
		}),
		LLMProvider: "test",
		Model:       "model",
	}
	output, err := usecase.Execute(context.Background(), GenerateDebateNameInput{Topic: "AI"})
	require.NoError(t, err)
	require.Equal(t, "Future of AI", output.Name)
}

func TestGenerateAgentsUsecase(t *testing.T) {
	t.Parallel()
	llm := &stubLLM{
		jsonResponse: `{"agents":[{"name":"Alex","stance":"Pro"},{"name":"Sam","stance":"Con"}]}`,
	}
	usecase := &GenerateAgentsUsecase{
		LLMResolver: contract.ResolverFunc[contract.LLM](func(name string) (contract.LLM, error) {
			return llm, nil
		}),
		LLMProvider: "test",
		Model:       "model",
	}
	output, err := usecase.Execute(context.Background(), GenerateAgentsInput{Topic: "AI", Count: 2})
	require.NoError(t, err)
	require.Len(t, output.Agents, 2)
	require.Equal(t, "Alex", output.Agents[0].Name)
	require.Equal(t, "Pro", output.Agents[0].Stance)
}
