import { useState, useCallback, useRef } from 'react';
import { useRevalidator } from 'react-router-dom';
import { useStoreActions } from '@/app/store/hooks';
import { createSSEStream } from '@/services/api/sse';

export interface UseSSEOptions {
    url: string;
    onRound: (round: { agent_id: string; content: string }) => void;
    onDone?: () => void;
}

export interface UseSSEReturn {
    isThinking: boolean;
    isStreaming: boolean;
    start: () => void;
    stop: () => void;
}

/**
 * Manages an SSE connection with manual start/stop controls.
 * Shows a 100ms inactivity "thinking" indicator between rounds.
 * Revalidates the route when the stream closes.
 */
export function useSSE(options: UseSSEOptions): UseSSEReturn {
    const [isThinking, setIsThinking] = useState(false);
    const [isStreaming, setIsStreaming] = useState(false);
    const callbacksRef = useRef(options);
    const thinkingTimerRef = useRef<number | null>(null);
    const closeStreamRef = useRef<(() => void) | null>(null);
    const revalidator = useRevalidator();
    const revalidatorRef = useRef(revalidator);
    const setSseActive = useStoreActions((store) => store.sessionUI.setSseActive);

    // Keep refs fresh without triggering effects
    // eslint-disable-next-line react-hooks/refs
    callbacksRef.current = options;
    // eslint-disable-next-line react-hooks/refs
    revalidatorRef.current = revalidator;

    const clearThinkingTimer = useCallback(() => {
        if (thinkingTimerRef.current !== null) {
            window.clearTimeout(thinkingTimerRef.current);
            thinkingTimerRef.current = null;
        }
    }, []);

    const stop = useCallback(() => {
        clearThinkingTimer();
        closeStreamRef.current?.();
        closeStreamRef.current = null;
        setIsThinking(false);
        setIsStreaming(false);
        setSseActive(false);
    }, [clearThinkingTimer, setSseActive]);

    const start = useCallback(() => {
        // Close any existing stream first
        stop();

        const url = callbacksRef.current.url;
        if (!url) return;

        let closed = false;
        setIsStreaming(true);
        setSseActive(true);

        // Show thinking skeleton if no event arrives within 100ms
        thinkingTimerRef.current = window.setTimeout(() => {
            if (!closed) setIsThinking(true);
        }, 100);

        const closeStream = createSSEStream({
            url,
            onRound: (round) => {
                if (closed) return;
                clearThinkingTimer();
                setIsThinking(false);
                callbacksRef.current.onRound(round);
                // Re-arm 100ms thinking timer for the gap between rounds
                thinkingTimerRef.current = window.setTimeout(() => {
                    if (!closed) setIsThinking(true);
                }, 100);
            },
            onDone: () => {
                if (closed) return;
                closed = true;
                clearThinkingTimer();
                setIsThinking(false);
                setIsStreaming(false);
                setSseActive(false);
                callbacksRef.current.onDone?.();
                revalidatorRef.current.revalidate();
            },
        });

        closeStreamRef.current = () => {
            closed = true;
            closeStream();
        };
    }, [stop, clearThinkingTimer, setSseActive]);

    return { isThinking, isStreaming, start, stop };
}
