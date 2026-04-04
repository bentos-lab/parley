export interface WhatsAppStatusResponse {
    connected: boolean;
}

export interface WhatsAppQREvent {
    code: string;
    timeout: number;
}

export interface WhatsAppScannedEvent {
    scanned: true;
}

export interface WhatsAppErrorEvent {
    error: string;
}

export type WhatsAppSSEEvent = WhatsAppQREvent | WhatsAppScannedEvent | WhatsAppErrorEvent;

export type WhatsAppConnectState =
    | { status: 'idle' }
    | { status: 'checking' }
    | { status: 'disconnected' }
    | { status: 'connecting' }
    | { status: 'qr'; code: string; expiresAt: number }
    | { status: 'scanned' }
    | { status: 'connected' }
    | { status: 'timeout' }
    | { status: 'error'; message: string };

export function isQREvent(event: WhatsAppSSEEvent): event is WhatsAppQREvent {
    return 'code' in event && 'timeout' in event;
}

export function isScannedEvent(event: WhatsAppSSEEvent): event is WhatsAppScannedEvent {
    return 'scanned' in event && event.scanned === true;
}

export function isErrorEvent(event: WhatsAppSSEEvent): event is WhatsAppErrorEvent {
    return 'error' in event && typeof event.error === 'string' && event.error.length > 0;
}
