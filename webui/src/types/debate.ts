export interface Agent {
    id: string;
    name: string;
    stance: string;
    voiceName: string;
}

export interface Round {
    agent_id: string;
    message: string;
}

export interface Debate {
    id: string;
    name: string;
    normalizedName: string;
    topic: string;
    agents: Agent[];
    rounds: Round[];
    ttsProvider: string;
}

export interface DebateSummary {
    id: string;
    name: string;
    topic: string;
}

export interface CreateDebateResponse {
    name: string;
    id: string;
}

export interface RoundResponse {
    agent_id: string;
    content: string;
}
