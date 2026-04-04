import { useCallback, useEffect, useRef, useState } from 'react';
import { connectWhatsAppSSE, disconnectWhatsApp, getWhatsAppStatus } from '@/services/api/whatsapp';
import { getErrorMessage } from '@/services/api/http';
import type { WhatsAppConnectState } from '@/types/whatsapp';

export interface UseWhatsAppConnectReturn {
    state: WhatsAppConnectState;
    /** Remaining seconds until QR expires (only valid in 'qr' state) */
    countdown: number;
    /** Start the connection flow */
    connect: () => void;
    /** Confirm synchronization is complete (closes SSE) */
    confirmSync: () => void;
    /** Disconnect existing session */
    disconnect: () => Promise<void>;
    /** Refresh the page to reload the current connection state */
    retry: () => void;
    /** Check connection status */
    checkStatus: () => Promise<void>;
}

/**
 * State machine hook for WhatsApp QR connection flow.
 * Manages: idle -> checking -> disconnected/connected
 *          disconnected -> connecting -> qr -> scanned -> connected
 *          qr -> timeout (on countdown expiry)
 *          any -> error (on failure)
 */
export function useWhatsAppConnect(): UseWhatsAppConnectReturn {
    const [state, setState] = useState<WhatsAppConnectState>({ status: 'idle' });
    const [countdown, setCountdown] = useState(0);

    const closeSSERef = useRef<(() => void) | null>(null);
    const countdownIntervalRef = useRef<number | null>(null);
    const expiresAtRef = useRef<number>(0);

    const cleanup = useCallback(() => {
        if (closeSSERef.current) {
            closeSSERef.current();
            closeSSERef.current = null;
        }

        if (countdownIntervalRef.current !== null) {
            window.clearInterval(countdownIntervalRef.current);
            countdownIntervalRef.current = null;
        }
    }, []);

    useEffect(() => cleanup, [cleanup]);

    const checkStatus = useCallback(async () => {
        setState({ status: 'checking' });

        try {
            const result = await getWhatsAppStatus();
            setState(result.connected ? { status: 'connected' } : { status: 'disconnected' });
        } catch (error) {
            setState({
                status: 'error',
                message: getErrorMessage(error, 'Failed to check connection status'),
            });
        }
    }, []);

    const startCountdown = useCallback(
        (expiresAt: number) => {
            expiresAtRef.current = expiresAt;

            const updateCountdown = () => {
                const remaining = Math.max(
                    0,
                    Math.ceil((expiresAtRef.current - Date.now()) / 1000),
                );
                setCountdown(remaining);

                if (remaining <= 0) {
                    cleanup();
                    setState({ status: 'timeout' });
                }
            };

            updateCountdown();
            countdownIntervalRef.current = window.setInterval(updateCountdown, 1000);
        },
        [cleanup],
    );

    const connect = useCallback(() => {
        cleanup();
        setCountdown(0);
        setState({ status: 'connecting' });

        closeSSERef.current = connectWhatsAppSSE({
            onQR: (code, timeout) => {
                const expiresAt = Date.now() + timeout;
                setState({ status: 'qr', code, expiresAt });
                startCountdown(expiresAt);
            },
            onScanned: () => {
                if (countdownIntervalRef.current !== null) {
                    window.clearInterval(countdownIntervalRef.current);
                    countdownIntervalRef.current = null;
                }
                setCountdown(0);
                setState({ status: 'scanned' });
            },
            onError: (error) => {
                cleanup();
                setState({
                    status: 'error',
                    message: getErrorMessage(error, 'Failed to connect WhatsApp'),
                });
            },
            onClose: () => {
                closeSSERef.current = null;
            },
        });
    }, [cleanup, startCountdown]);

    const confirmSync = useCallback(() => {
        cleanup();
        setCountdown(0);
        setState({ status: 'connected' });
    }, [cleanup]);

    const disconnect = useCallback(async () => {
        cleanup();
        setState({ status: 'checking' });

        try {
            await disconnectWhatsApp();
            setCountdown(0);
            setState({ status: 'disconnected' });
        } catch (error) {
            setState({ status: 'error', message: getErrorMessage(error, 'Failed to disconnect') });
        }
    }, [cleanup]);

    const retry = useCallback(() => {
        cleanup();
        window.location.reload();
    }, [cleanup]);

    return {
        state,
        countdown,
        connect,
        confirmSync,
        disconnect,
        retry,
        checkStatus,
    };
}
