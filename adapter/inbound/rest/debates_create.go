package rest

import (
	"encoding/json"
	"net/http"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
)

type createDebateRequest struct {
	Name        string               `json:"name"`
	Topic       string               `json:"topic"`
	Agents      []debate.DebateAgent `json:"agents"`
	NumAgents   int                  `json:"num_agents"`
	TTSProvider string               `json:"tts_provider"`
	AgentVoices map[string]string    `json:"agent_voices"`
	LLMProvider string               `json:"llm_provider"`
	LLMModel    string               `json:"llm_model"`
}

type createDebateResponse struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// createDebate handles creating a new debate.
// Parameters: w is the response writer, r is the HTTP request.
// Returns: nothing.
func (h *Handler) createDebate(w http.ResponseWriter, r *http.Request) {
	var req createDebateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.Agents) > 0 && req.NumAgents > 0 {
		writeError(w, http.StatusBadRequest, "agents and num_agents are mutually exclusive")
		return
	}
	usecases, cfg, ok := h.loadUsecases(w)
	if !ok {
		return
	}
	if req.TTSProvider == "" {
		req.TTSProvider = cfg.TTSProvider
	}
	ctx := core.WithLLMSelection(r.Context(), req.LLMProvider, req.LLMModel)
	name := req.Name
	if name == "" {
		if usecases.GenerateDebateName == nil {
			writeError(w, http.StatusBadRequest, "name generator is required")
			return
		}
		nameOutput, err := usecases.GenerateDebateName.Execute(ctx, core.GenerateDebateNameInput{Topic: req.Topic})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		name = nameOutput.Name
	}
	agents := req.Agents
	if len(agents) == 0 {
		if req.NumAgents <= 0 {
			writeError(w, http.StatusBadRequest, "num_agents must be greater than zero")
			return
		}
		if usecases.GenerateDebateAgents == nil {
			writeError(w, http.StatusBadRequest, "agent generator is required")
			return
		}
		agentsOutput, err := usecases.GenerateDebateAgents.Execute(ctx, core.GenerateAgentsInput{
			Topic: req.Topic,
			Count: req.NumAgents,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		agents = agentsOutput.Agents
	}
	if usecases.AssignDebateVoices == nil {
		writeError(w, http.StatusBadRequest, "voice assignment usecase is required")
		return
	}
	voicesOutput, err := usecases.AssignDebateVoices.Execute(ctx, core.AssignDebateVoicesInput{
		Agents:      agents,
		TTSProvider: req.TTSProvider,
		AgentVoices: req.AgentVoices,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	agents = voicesOutput.Agents
	output, err := usecases.CreateDebate.Execute(ctx, core.CreateDebateInput{
		Name:        name,
		Topic:       req.Topic,
		Agents:      agents,
		TTSProvider: req.TTSProvider,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, createDebateResponse{
		Name: output.Debate.Name,
		ID:   debate.IDFromFilename(output.Filename),
	})
}
