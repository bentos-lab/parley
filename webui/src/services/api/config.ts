import { API_PREFIX, request } from './http';
import { ConfigResponseSchema } from './schemas';
import type { ConfigResponse, ConfigUpdatePayload } from '@/types';

/**
 * Fetches the current runtime configuration from GET /api/config.
 */
export async function getConfig(): Promise<ConfigResponse> {
    const path = `${API_PREFIX}/config`;
    const response = await request<unknown>(path, { method: 'GET' });
    return ConfigResponseSchema.parse(response);
}

/**
 * Updates the persisted configuration via PUT /api/config.
 * Only provided keys are written.
 */
export async function updateConfig(payload: ConfigUpdatePayload): Promise<ConfigResponse> {
    const path = `${API_PREFIX}/config`;
    const response = await request<unknown>(path, {
        method: 'PUT',
        body: JSON.stringify(payload),
    });
    return ConfigResponseSchema.parse(response);
}
