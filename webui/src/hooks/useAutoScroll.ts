import { useCallback, useEffect, useRef, useState, type RefObject } from 'react';

export interface UseAutoScrollOptions {
    containerRef: RefObject<HTMLDivElement | null>;
    trigger: unknown;
}

export interface UseAutoScrollResult {
    paused: boolean;
    resumeNow: () => void;
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

        function handleWheel(event: WheelEvent) {
            // Only pause when scrolling up (user intentionally looking at older messages)
            if (event.deltaY < 0) {
                pausedRef.current = true;
                setPausedState(true);

                // Clear any existing timer - we stay paused until user clicks "Scroll to bottom"
                if (resumeTimer.current) {
                    clearTimeout(resumeTimer.current);
                    resumeTimer.current = null;
                }
            }
        }

        element.addEventListener('wheel', handleWheel, { passive: true });

        return () => {
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
