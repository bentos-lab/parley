export type {
    Agent,
    Round,
    Debate,
    DebateSummary,
    CreateDebateResponse,
    RoundResponse,
    OpenAIConfig,
    LLMConfig,
    InworldConfig,
    TTSConfig,
    ConfigResponse,
    ConfigUpdatePayload,
} from './debate';

export type {
    WhatsAppConnectState,
    WhatsAppErrorEvent,
    WhatsAppQREvent,
    WhatsAppScannedEvent,
    WhatsAppSSEEvent,
    WhatsAppStatusResponse,
} from './whatsapp';
export { isErrorEvent, isQREvent, isScannedEvent } from './whatsapp';
