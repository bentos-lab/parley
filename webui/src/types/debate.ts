export interface Agent {
    id: string;
    name: string;
    stance: string;
    voiceName: string;
}

export interface Round {
    agent_id: string;
    message: string;
    weakness?: string;
    new_point?: string;
    rebuttal?: string;
    summary?: string;
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
    weakness?: string;
    new_point?: string;
    rebuttal?: string;
    summary?: string;
}

// Config types matching GET/PUT /api/config
export interface OpenAIConfig {
    base_url: string;
    api_key: string;
    model: string;
}

export interface LLMConfig {
    provider: string;
    openai: OpenAIConfig;
}

export interface InworldConfig {
    api_key: string;
    model: string;
}

export interface TTSConfig {
    provider: string;
    inworld: InworldConfig;
}

export interface ConfigResponse {
    llm: LLMConfig;
    tts: TTSConfig;
}

export interface ConfigUpdatePayload {
    llm?: Partial<LLMConfig> & { openai?: Partial<OpenAIConfig> };
    tts?: Partial<TTSConfig> & { inworld?: Partial<InworldConfig> };
}
