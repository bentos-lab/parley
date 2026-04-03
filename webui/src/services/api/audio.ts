import { API_PREFIX, BASE, requestBlob } from './http';

/** Get audio for a specific round as a blob */
export function getRoundAudio(debateId: string, roundIndex: number): Promise<Blob> {
    return requestBlob(
        `${API_PREFIX}/debates/${encodeURIComponent(debateId)}/rounds/${roundIndex}/audio`,
    );
}

/** Build the URL for a round's audio (useful for <audio> src) */
export function roundAudioUrl(debateId: string, roundIndex: number): string {
    return `${BASE}${API_PREFIX}/debates/${encodeURIComponent(debateId)}/rounds/${roundIndex}/audio`;
}

/** Get full debate audio as a blob */
export function getDebateAudio(debateId: string): Promise<Blob> {
    return requestBlob(`${API_PREFIX}/debates/${encodeURIComponent(debateId)}/audio`);
}

/** Build the URL for full debate audio */
export function debateAudioUrl(debateId: string): string {
    return `${BASE}${API_PREFIX}/debates/${encodeURIComponent(debateId)}/audio`;
}
