import { useState, useCallback, useRef, useEffect } from 'react';
import { useRevalidator } from 'react-router-dom';
import { appendRound } from '@/services/api/rounds';
import type { Round, RoundResponse } from '@/types';

export interface UseRoundGenerationOptions {
    debateId: string;
    onRound: (round: Round, roundIndex: number) => void;
    onError?: (error: Error) => void;
}

export interface UseRoundGenerationReturn {
    /** True when waiting for round generation response */
    isGenerating: boolean;
    /** True when generation loop is active (generating or between rounds) */
    isActive: boolean;
    /** True when stop was requested and waiting for cleanup */
    isStopping: boolean;
    /** Number of rounds generated in current session */
    generatedCount: number;
    /** Index of the last generated round (0-based) */
    lastGeneratedRoundIndex: number | null;
    /** Start generating the specified number of rounds */
    start: (count: number, startingRoundIndex: number) => void;
    /** Stop generation after current round completes */
    stop: () => void;
}

/**
 * Manages sequential round generation via POST /api/debates/{id}/rounds.
 * Revalidates the route when generation completes or is stopped.
 * Audio playback is handled separately via useRoundAudioCache.
 */
export function useRoundGeneration(options: UseRoundGenerationOptions): UseRoundGenerationReturn {
    const [isGenerating, setIsGenerating] = useState(false);
    const [isActive, setIsActive] = useState(false);
    const [isStopping, setIsStopping] = useState(false);
    const [generatedCount, setGeneratedCount] = useState(0);
    const [lastGeneratedRoundIndex, setLastGeneratedRoundIndex] = useState<number | null>(null);

    const optionsRef = useRef(options);
    const stopRequestedRef = useRef(false);
    const isMountedRef = useRef(true);
    const revalidator = useRevalidator();
    const revalidatorRef = useRef(revalidator);

    const setIsGeneratingIfMounted = useCallback((value: boolean) => {
        if (isMountedRef.current) {
            setIsGenerating(value);
        }
    }, []);

    const setIsActiveIfMounted = useCallback((value: boolean) => {
        if (isMountedRef.current) {
            setIsActive(value);
        }
    }, []);

    const setIsStoppingIfMounted = useCallback((value: boolean) => {
        if (isMountedRef.current) {
            setIsStopping(value);
        }
    }, []);

    const setGeneratedCountIfMounted = useCallback((value: number | ((prev: number) => number)) => {
        if (isMountedRef.current) {
            setGeneratedCount(value);
        }
    }, []);

    const setLastGeneratedRoundIndexIfMounted = useCallback((value: number | null) => {
        if (isMountedRef.current) {
            setLastGeneratedRoundIndex(value);
        }
    }, []);

    // Keep refs fresh using useEffect
    useEffect(() => {
        optionsRef.current = options;
    });
    useEffect(() => {
        revalidatorRef.current = revalidator;
    });
    useEffect(() => {
        isMountedRef.current = true;

        return () => {
            isMountedRef.current = false;
            stopRequestedRef.current = true;
        };
    }, []);

    const stop = useCallback(() => {
        if (!isActive || isStopping) return;

        setIsStoppingIfMounted(true);
        stopRequestedRef.current = true;
    }, [isActive, isStopping, setIsStoppingIfMounted]);

    const start = useCallback(
        async (count: number, startingRoundIndex: number) => {
            if (isActive) return;

            stopRequestedRef.current = false;
            setIsActiveIfMounted(true);
            setGeneratedCountIfMounted(0);
            setLastGeneratedRoundIndexIfMounted(null);

            const { debateId, onRound, onError } = optionsRef.current;

            // Track the current round index
            let currentRoundIndex = startingRoundIndex;

            for (let i = 0; i < count; i++) {
                if (stopRequestedRef.current) break;

                setIsGeneratingIfMounted(true);

                try {
                    const response: RoundResponse = await appendRound(debateId);

                    if (stopRequestedRef.current || !isMountedRef.current) {
                        setIsGeneratingIfMounted(false);
                        break;
                    }

                    // Convert RoundResponse to Round format
                    const round: Round = {
                        agent_id: response.agent_id,
                        message: response.content,
                        weakness: response.weakness,
                        new_point: response.new_point,
                        rebuttal: response.rebuttal,
                        summary: response.summary,
                    };

                    const roundIndex = currentRoundIndex;
                    currentRoundIndex++;

                    onRound(round, roundIndex);
                    setGeneratedCountIfMounted((prev) => prev + 1);
                    setLastGeneratedRoundIndexIfMounted(roundIndex);
                    setIsGeneratingIfMounted(false);
                } catch (err) {
                    setIsGeneratingIfMounted(false);

                    if (isMountedRef.current) {
                        onError?.(err instanceof Error ? err : new Error(String(err)));
                    }

                    break;
                }
            }

            setIsActiveIfMounted(false);
            setIsGeneratingIfMounted(false);
            setIsStoppingIfMounted(false);

            // Revalidate to sync with server state
            if (isMountedRef.current) {
                revalidatorRef.current.revalidate();
            }
        },
        [
            isActive,
            setGeneratedCountIfMounted,
            setIsActiveIfMounted,
            setIsGeneratingIfMounted,
            setIsStoppingIfMounted,
            setLastGeneratedRoundIndexIfMounted,
        ],
    );

    return {
        isGenerating,
        isActive,
        isStopping,
        generatedCount,
        lastGeneratedRoundIndex,
        start,
        stop,
    };
}
