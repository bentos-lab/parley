import { useCallback, useEffect, useRef, useState, type RefObject } from 'react';

export interface UseAutoScrollOptions {
    containerRef: RefObject<HTMLDivElement | null>;
    trigger: unknown;
}

export interface UseAutoScrollResult {
    paused: boolean;
    resumeNow: () => void;
}

const BOTTOM_THRESHOLD_PX = 24;

function isNearBottom(element: HTMLDivElement | null): boolean {
    if (!element) {
        return true;
    }

    return element.scrollHeight - element.scrollTop - element.clientHeight <= BOTTOM_THRESHOLD_PX;
}

export function useAutoScroll({
    containerRef,
    trigger,
}: UseAutoScrollOptions): UseAutoScrollResult {
    const pausedRef = useRef(false);
    const resumeTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
    const [pausedState, setPausedState] = useState(false);

    const resumeNow = useCallback(() => {
        pausedRef.current = false;
        setPausedState(false);

        if (resumeTimer.current) {
            clearTimeout(resumeTimer.current);
            resumeTimer.current = null;
        }

        if (containerRef.current) {
            containerRef.current.scrollTop = containerRef.current.scrollHeight;
        }
    }, [containerRef]);

    useEffect(() => {
        if (!pausedRef.current && containerRef.current) {
            containerRef.current.scrollTop = containerRef.current.scrollHeight;
        }
    }, [containerRef, trigger]);

    useEffect(() => {
        const element = containerRef.current;

        if (!element) {
            return;
        }

        function syncPausedState() {
            const nextPaused = !isNearBottom(element);
            pausedRef.current = nextPaused;
            setPausedState(nextPaused);

            if (!nextPaused && resumeTimer.current) {
                clearTimeout(resumeTimer.current);
                resumeTimer.current = null;
            }
        }

        function handleScroll() {
            syncPausedState();
        }

        function handleWheel(event: WheelEvent) {
            if (event.deltaY < 0) {
                pausedRef.current = true;
                setPausedState(true);

                if (resumeTimer.current) {
                    clearTimeout(resumeTimer.current);
                    resumeTimer.current = null;
                }
                return;
            }

            requestAnimationFrame(syncPausedState);
        }

        syncPausedState();
        element.addEventListener('scroll', handleScroll, { passive: true });
        element.addEventListener('wheel', handleWheel, { passive: true });

        return () => {
            element.removeEventListener('scroll', handleScroll);
            element.removeEventListener('wheel', handleWheel);

            if (resumeTimer.current) {
                clearTimeout(resumeTimer.current);
                resumeTimer.current = null;
            }
        };
    }, [containerRef, resumeNow]);

    return {
        paused: pausedState,
        resumeNow,
    };
}
