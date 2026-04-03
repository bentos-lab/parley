import { API_PREFIX, request } from './http';
import type { RoundResponse } from '@/types';

export interface AppendRoundPayload {
    agent_id?: string;
    content: string;
}

export function appendRound(debateId: string, payload: AppendRoundPayload): Promise<RoundResponse> {
    const path = `${API_PREFIX}/debates/${encodeURIComponent(debateId)}/rounds`;
    return request<RoundResponse>(path, {
        method: 'POST',
        body: JSON.stringify(payload),
    });
}
