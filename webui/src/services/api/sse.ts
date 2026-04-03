import { BASE } from './http';

export interface SSEOptions {
    url: string;
    onRound: (round: { agent_id: string; content: string }) => void;
    onDone: () => void;
}

/**
 * Opens an EventSource SSE connection and returns a cleanup function.
 * Each SSE event is a complete turn: data: {"agent_id":"...","content":"..."}
 * The stream closes without a [DONE] sentinel — onerror fires on clean close.
 */
export function createSSEStream({ url, onRound, onDone }: SSEOptions): () => void {
    const fullUrl = url.startsWith('http') ? url : `${BASE}${url}`;
    const es = new EventSource(fullUrl);

    es.onmessage = (e: MessageEvent<string>) => {
        try {
            const round = JSON.parse(e.data) as { agent_id: string; content: string };
            onRound(round);
        } catch {
            // ignore malformed events
        }
    };

    es.onerror = () => {
        es.close();
        onDone();
    };

    return () => es.close();
}
