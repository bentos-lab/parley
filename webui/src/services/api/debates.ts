import { API_PREFIX, request, requestVoid } from './http';
import type { Debate, DebateSummary, CreateDebateResponse } from '@/types';
import type { RawDebateResponse } from './schemas';

const DEBATES_ENDPOINT = `${API_PREFIX}/debates`;

export function listDebates(): Promise<DebateSummary[]> {
    return request<DebateSummary[]>(DEBATES_ENDPOINT);
}

export function getDebate(id: string): Promise<RawDebateResponse> {
    return request<RawDebateResponse>(`${DEBATES_ENDPOINT}/${encodeURIComponent(id)}`);
}

export interface CreateDebatePayload {
    topic: string;
    name?: string;
    tts_provider?: string;
    agents?: Array<{ name?: string; stance?: string }>;
}

export function createDebate(payload: CreateDebatePayload): Promise<CreateDebateResponse> {
    return request<CreateDebateResponse>(DEBATES_ENDPOINT, {
        method: 'POST',
        body: JSON.stringify(payload),
    });
}

export function updateDebate(id: string, debate: Debate): Promise<void> {
    const payload = {
        id: debate.id,
        name: debate.name,
        normalized_name: debate.normalizedName,
        topic: debate.topic,
        tts_provider: debate.ttsProvider,
        agents: debate.agents.map((a) => ({
            id: a.id,
            name: a.name,
            stance: a.stance,
            voice_name: a.voiceName,
        })),
        rounds: debate.rounds,
    };

    return requestVoid(`${DEBATES_ENDPOINT}/${encodeURIComponent(id)}`, {
        method: 'PUT',
        body: JSON.stringify(payload),
    });
}

export function deleteDebate(id: string): Promise<void> {
    return requestVoid(`${DEBATES_ENDPOINT}/${encodeURIComponent(id)}`, { method: 'DELETE' });
}
