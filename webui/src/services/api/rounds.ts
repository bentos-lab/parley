import { API_PREFIX, request } from './http';
import { RoundResponseSchema } from './schemas';
import type { RoundResponse } from '@/types';

export interface AppendRoundPayload {
    agent_id?: string;
    content?: string;
}

export async function appendRound(
    debateId: string,
    payload?: AppendRoundPayload,
): Promise<RoundResponse> {
    const path = `${API_PREFIX}/debates/${encodeURIComponent(debateId)}/rounds`;
    const response = await request<unknown>(path, {
        method: 'POST',
        body: JSON.stringify(payload ?? {}),
    });
    return RoundResponseSchema.parse(response);
}

export async function getRound(debateId: string, index: number): Promise<RoundResponse> {
    const path = `${API_PREFIX}/debates/${encodeURIComponent(debateId)}/rounds/${index}`;
    const response = await request<unknown>(path, { method: 'GET' });
    return RoundResponseSchema.parse(response);
}
