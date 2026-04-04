export { listDebates, getDebate, createDebate, updateDebate, deleteDebate } from './debates';
export type { CreateDebatePayload } from './debates';
export { appendRound, getRound } from './rounds';
export type { AppendRoundPayload } from './rounds';
export { getRoundAudio, roundAudioUrl, getDebateAudio, debateAudioUrl } from './audio';
export { getConfig, updateConfig } from './config';
export { connectWhatsAppSSE, disconnectWhatsApp, getWhatsAppStatus } from './whatsapp';
export type { WhatsAppSSEOptions } from './whatsapp';
// SSE exports are deprecated - use POST /api/debates/{id}/rounds for sequential generation
// export { createSSEStream } from './sse';
// export type { SSEOptions } from './sse';
export { ApiError, BASE } from './http';
