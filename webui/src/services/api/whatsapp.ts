import { API_PREFIX, ApiError, BASE, request, requestVoid } from './http';
import type { WhatsAppSSEEvent, WhatsAppStatusResponse } from '@/types/whatsapp';

const WHATSAPP_ENDPOINT = `${API_PREFIX}/connect/whatsapp`;

export function getWhatsAppStatus(): Promise<WhatsAppStatusResponse> {
    return request<WhatsAppStatusResponse>(WHATSAPP_ENDPOINT, { method: 'GET' });
}

export function disconnectWhatsApp(): Promise<void> {
    return requestVoid(WHATSAPP_ENDPOINT, { method: 'DELETE' });
}

export interface WhatsAppSSEOptions {
    onQR: (code: string, timeout: number) => void;
    onScanned: () => void;
    onError: (error: Error) => void;
    onClose: () => void;
}

export function connectWhatsAppSSE(options: WhatsAppSSEOptions): () => void {
    const url = `${BASE}${WHATSAPP_ENDPOINT}?connect=true`;
    const es = new EventSource(url);

    es.onmessage = (event: MessageEvent<string>) => {
        try {
            const payload = JSON.parse(event.data) as WhatsAppSSEEvent;

            if ('code' in payload && 'timeout' in payload) {
                options.onQR(payload.code, payload.timeout);
                return;
            }

            if ('scanned' in payload && payload.scanned === true) {
                options.onScanned();
                return;
            }

            if ('error' in payload && typeof payload.error === 'string') {
                es.close();
                options.onError(new ApiError(500, payload.error, payload));
            }
        } catch {
            // Ignore malformed events.
        }
    };

    es.onerror = () => {
        if (es.readyState === EventSource.CLOSED) {
            es.close();
            options.onClose();
            return;
        }

        es.close();
        options.onError(new Error('Connection failed'));
    };

    return () => {
        es.close();
    };
}
