package rest

type roundRequest struct {
	AgentID string `json:"agent_id"`
	Content string `json:"content"`
}

type roundResponse struct {
	AgentID  string `json:"agent_id"`
	Content  string `json:"content"`
	Weakness string `json:"weakness"`
	NewPoint string `json:"new_point"`
	Rebuttal string `json:"rebuttal"`
	Summary  string `json:"summary"`
}
