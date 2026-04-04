import { useCallback, useSyncExternalStore } from 'react';
import { getRoundAudio } from '@/services/api/audio';
import { getErrorMessage } from '@/services/api/http';

// ---------------------------------------------------------------------------
// Global audio cache (persists across component instances and routes)
// ---------------------------------------------------------------------------

type AudioStatus = 'idle' | 'loading' | 'ready' | 'playing' | 'error';

interface CacheEntry {
    url: string | null;
    status: AudioStatus;
    duration: number | null;
    currentTime: number;
    errorMessage?: string;
}

const cache = new Map<string, CacheEntry>();
const inFlightLoads = new Map<string, Promise<string | null>>();
const listeners = new Set<() => void>();

function getCacheKey(debateId: string, roundIndex: number): string {
    return `${debateId}:${roundIndex}`;
}

function notifyListeners() {
    listeners.forEach((fn) => fn());
}

function getEntry(key: string): CacheEntry {
    return (
        cache.get(key) ?? {
            url: null,
            status: 'idle',
            duration: null,
            currentTime: 0,
            errorMessage: undefined,
        }
    );
}

function isErrorStatusForCurrentItem(item: QueueItem): boolean {
    const key = getCacheKey(item.debateId, item.roundIndex);
    return getEntry(key).status === 'error';
}

function setEntry(key: string, partial: Partial<CacheEntry>) {
    const current = getEntry(key);
    cache.set(key, { ...current, ...partial });
    notifyListeners();
}

// ---------------------------------------------------------------------------
// Queue State (sequential playback)
// ---------------------------------------------------------------------------

export type PlayerStatus = 'idle' | 'loading' | 'playing' | 'paused';

export interface QueueItem {
    debateId: string;
    roundIndex: number;
}

interface QueueState {
    /** Items waiting to be played */
    queue: QueueItem[];
    /** Currently playing/paused item (null if idle) */
    current: QueueItem | null;
    /** Items that have finished playing in this session */
    history: QueueItem[];
    /** Overall player status */
    playerStatus: PlayerStatus;
}

let queueState: QueueState = {
    queue: [],
    current: null,
    history: [],
    playerStatus: 'idle',
};

function setQueueState(partial: Partial<QueueState>) {
    queueState = { ...queueState, ...partial };
    notifyListeners();
}

function getQueueState(): QueueState {
    return queueState;
}

// ---------------------------------------------------------------------------
// Global audio element for playback (singleton)
// ---------------------------------------------------------------------------

let globalAudio: HTMLAudioElement | null = null;

function getGlobalAudio(): HTMLAudioElement {
    if (!globalAudio) {
        globalAudio = new Audio();
        globalAudio.addEventListener('ended', handleAudioEnded);
        globalAudio.addEventListener('loadedmetadata', () => {
            const { current } = getQueueState();
            if (current && globalAudio && !isNaN(globalAudio.duration)) {
                const key = getCacheKey(current.debateId, current.roundIndex);
                setEntry(key, { duration: globalAudio.duration });
            }
        });
        globalAudio.addEventListener('timeupdate', () => {
            const { current } = getQueueState();
            if (current && globalAudio) {
                const key = getCacheKey(current.debateId, current.roundIndex);
                setEntry(key, { currentTime: globalAudio.currentTime });
            }
        });
    }
    return globalAudio;
}

/** Called when audio ends — advance to next in queue */
function handleAudioEnded() {
    const { current, queue, history } = getQueueState();
    if (current) {
        const key = getCacheKey(current.debateId, current.roundIndex);
        const currentEntry = getEntry(key);
        setEntry(key, {
            status: currentEntry.status === 'error' ? 'error' : 'ready',
            currentTime: 0,
        });

        // Move current to history
        const newHistory = [...history, current];

        if (queue.length > 0) {
            // Play next item
            const [next, ...rest] = queue;
            setQueueState({
                current: next,
                queue: rest,
                history: newHistory,
                playerStatus: 'loading',
            });
            void playCurrentItem(next);
        } else {
            // Queue exhausted
            setQueueState({
                current: null,
                queue: [],
                history: newHistory,
                playerStatus: 'idle',
            });
        }
    }
}

/** Internal: play a specific item (assumes it's now current) */
async function playCurrentItem(item: QueueItem): Promise<void> {
    const key = getCacheKey(item.debateId, item.roundIndex);
    const audio = getGlobalAudio();

    // Load if needed
    const url = await loadRoundAudio(item.debateId, item.roundIndex);
    if (!url) {
        // Error loading — skip to next
        handleAudioEnded();
        return;
    }

    // Check we're still supposed to play this item
    if (
        getQueueState().current?.debateId !== item.debateId ||
        getQueueState().current?.roundIndex !== item.roundIndex
    ) {
        return; // Playback was changed while loading
    }

    if (audio.src !== url) {
        audio.src = url;
        audio.load();
    }

    try {
        await audio.play();
        setEntry(key, { status: 'playing' });
        setQueueState({ playerStatus: 'playing' });
    } catch {
        setEntry(key, { status: 'error' });
        // Skip to next on error
        handleAudioEnded();
    }
}

// ---------------------------------------------------------------------------
// Load round audio into cache
// ---------------------------------------------------------------------------

async function loadRoundAudio(debateId: string, roundIndex: number): Promise<string | null> {
    const key = getCacheKey(debateId, roundIndex);
    const entry = getEntry(key);

    // Already loaded
    if (entry.url) {
        return entry.url;
    }

    // Already loading — wait for existing request
    const existing = inFlightLoads.get(key);
    if (existing) {
        return existing;
    }

    setEntry(key, { status: 'loading' });

    const loadPromise = getRoundAudio(debateId, roundIndex)
        .then((blob) => {
            const url = URL.createObjectURL(blob);
            setEntry(key, { url, status: 'ready', errorMessage: undefined });
            return url;
        })
        .catch((error) => {
            setEntry(key, {
                status: 'error',
                errorMessage: getErrorMessage(error, 'Something went wrong'),
            });
            return null;
        })
        .finally(() => {
            inFlightLoads.delete(key);
        });

    inFlightLoads.set(key, loadPromise);
    return loadPromise;
}

// ---------------------------------------------------------------------------
// Queue controls
// ---------------------------------------------------------------------------

/** Enqueue a round for autoplay (adds to end of queue) */
function enqueueRound(debateId: string, roundIndex: number): void {
    const { queue, current, playerStatus } = getQueueState();
    const item: QueueItem = { debateId, roundIndex };

    // Check if already queued or playing
    const isQueued = queue.some((q) => q.debateId === debateId && q.roundIndex === roundIndex);
    const isCurrent = current?.debateId === debateId && current?.roundIndex === roundIndex;
    if (isQueued || isCurrent) return;

    if (!current && playerStatus === 'idle') {
        // Nothing playing — start immediately
        setQueueState({ current: item, playerStatus: 'loading' });
        void playCurrentItem(item);
    } else {
        // Add to queue
        setQueueState({ queue: [...queue, item] });
    }
}

/** Play a specific round immediately (interrupts current, clears queue) */
async function playRoundNow(debateId: string, roundIndex: number): Promise<void> {
    const audio = getGlobalAudio();
    const { current } = getQueueState();

    // Stop current audio
    if (current) {
        audio.pause();
        const key = getCacheKey(current.debateId, current.roundIndex);
        setEntry(key, { status: 'ready' });
    }

    const item: QueueItem = { debateId, roundIndex };
    setQueueState({
        current: item,
        queue: [],
        playerStatus: 'loading',
    });

    await playCurrentItem(item);
}

/** Toggle play/pause for current item */
function togglePlayPause(): void {
    const audio = getGlobalAudio();
    const { current, playerStatus } = getQueueState();

    if (!current) return;

    const key = getCacheKey(current.debateId, current.roundIndex);

    if (playerStatus === 'playing') {
        audio.pause();
        setEntry(key, { status: 'ready' });
        setQueueState({ playerStatus: 'paused' });
    } else if (playerStatus === 'paused') {
        void audio
            .play()
            .then(() => {
                setEntry(key, { status: 'playing' });
                setQueueState({ playerStatus: 'playing' });
            })
            .catch(() => {
                setEntry(key, { status: 'error' });

                if (isErrorStatusForCurrentItem(current)) {
                    handleAudioEnded();
                }
            });
    }
}

/** Skip to next item in queue */
function skipNext(): void {
    const audio = getGlobalAudio();
    const { current, queue, history } = getQueueState();

    if (!current) return;

    audio.pause();
    const key = getCacheKey(current.debateId, current.roundIndex);
    setEntry(key, { status: 'ready', currentTime: 0 });

    const newHistory = [...history, current];

    if (queue.length > 0) {
        const [next, ...rest] = queue;
        setQueueState({
            current: next,
            queue: rest,
            history: newHistory,
            playerStatus: 'loading',
        });
        void playCurrentItem(next);
    } else {
        setQueueState({
            current: null,
            queue: [],
            history: newHistory,
            playerStatus: 'idle',
        });
    }
}

/** Go back to previous item in history */
function skipPrevious(): void {
    const audio = getGlobalAudio();
    const { current, queue, history } = getQueueState();

    if (history.length === 0) {
        if (!current) {
            return;
        }

        if (current.roundIndex <= 0) {
            // No history and nothing before this round — restart current
            if (globalAudio) {
                globalAudio.currentTime = 0;
            }
            return;
        }

        audio.pause();
        const key = getCacheKey(current.debateId, current.roundIndex);
        setEntry(key, { status: 'ready', currentTime: 0 });

        const previousItem: QueueItem = {
            debateId: current.debateId,
            roundIndex: current.roundIndex - 1,
        };

        setQueueState({
            current: previousItem,
            queue: [current, ...queue],
            playerStatus: 'loading',
        });
        void playCurrentItem(previousItem);
        return;
    }

    // Pause current
    if (current) {
        audio.pause();
        const key = getCacheKey(current.debateId, current.roundIndex);
        setEntry(key, { status: 'ready', currentTime: 0 });
    }

    // Move last history item to current, push current to front of queue
    const prev = history[history.length - 1];
    const newHistory = history.slice(0, -1);
    const newQueue = current ? [current, ...queue] : queue;

    setQueueState({
        current: prev,
        queue: newQueue,
        history: newHistory,
        playerStatus: 'loading',
    });
    void playCurrentItem(prev);
}

/** Stop all playback and clear queue */
function stopAll(): void {
    const audio = getGlobalAudio();
    const { current } = getQueueState();

    audio.pause();

    if (current) {
        const key = getCacheKey(current.debateId, current.roundIndex);
        setEntry(key, { status: 'ready', currentTime: 0 });
    }

    setQueueState({
        current: null,
        queue: [],
        history: [],
        playerStatus: 'idle',
    });
}

/** Reset a specific round (restart from beginning) */
function resetRound(debateId: string, roundIndex: number): void {
    const key = getCacheKey(debateId, roundIndex);
    const entry = getEntry(key);
    const audio = getGlobalAudio();
    const { current } = getQueueState();

    // If this round is currently playing, restart it
    if (current?.debateId === debateId && current?.roundIndex === roundIndex) {
        audio.currentTime = 0;
        setEntry(key, { currentTime: 0 });
        return;
    }

    // Reset currentTime in cache (keep the audio loaded)
    if (entry.url) {
        setEntry(key, { currentTime: 0 });
    }
}

// ---------------------------------------------------------------------------
// React hook
// ---------------------------------------------------------------------------

function subscribe(callback: () => void): () => void {
    listeners.add(callback);
    return () => listeners.delete(callback);
}

function getSnapshot(): Map<string, CacheEntry> {
    return cache;
}

export interface RoundAudioState {
    status: AudioStatus;
    duration: number | null;
    currentTime: number;
    errorMessage?: string;
}

export interface AudioQueueState {
    /** Items waiting to be played */
    queue: QueueItem[];
    /** Currently playing/paused item (null if idle) */
    current: QueueItem | null;
    /** Items that have finished playing */
    history: QueueItem[];
    /** Overall player status */
    playerStatus: PlayerStatus;
    /** Total items in queue + current */
    totalQueued: number;
    /** Position in queue (1-indexed, 0 if nothing playing) */
    currentPosition: number;
}

export function useRoundAudioCache() {
    // Subscribe to cache changes
    useSyncExternalStore(subscribe, getSnapshot, getSnapshot);

    const getStatus = useCallback((debateId: string, roundIndex: number): RoundAudioState => {
        const key = getCacheKey(debateId, roundIndex);
        const entry = getEntry(key);
        return {
            status: entry.status,
            duration: entry.duration,
            currentTime: entry.currentTime,
            errorMessage: entry.errorMessage,
        };
    }, []);

    /** @deprecated Use enqueue() for autoplay, playNow() for immediate */
    const play = useCallback(async (debateId: string, roundIndex: number) => {
        await playRoundNow(debateId, roundIndex);
    }, []);

    const playNow = useCallback(async (debateId: string, roundIndex: number) => {
        await playRoundNow(debateId, roundIndex);
    }, []);

    const enqueue = useCallback((debateId: string, roundIndex: number) => {
        enqueueRound(debateId, roundIndex);
    }, []);

    const load = useCallback(async (debateId: string, roundIndex: number) => {
        return loadRoundAudio(debateId, roundIndex);
    }, []);

    const stop = useCallback(() => {
        stopAll();
    }, []);

    const reset = useCallback((debateId: string, roundIndex: number) => {
        resetRound(debateId, roundIndex);
    }, []);

    const togglePause = useCallback(() => {
        togglePlayPause();
    }, []);

    const next = useCallback(() => {
        skipNext();
    }, []);

    const previous = useCallback(() => {
        skipPrevious();
    }, []);

    return {
        getStatus,
        play,
        playNow,
        enqueue,
        load,
        stop,
        reset,
        togglePause,
        next,
        previous,
    };
}

// ---------------------------------------------------------------------------
// Hook for queue state (player controls)
// ---------------------------------------------------------------------------

const IDLE_QUEUE_STATE: QueueState = {
    queue: [],
    current: null,
    history: [],
    playerStatus: 'idle',
};

export function useAudioQueueState(): AudioQueueState {
    const state = useSyncExternalStore(subscribe, getQueueState, () => IDLE_QUEUE_STATE);

    const totalQueued = (state.current ? 1 : 0) + state.queue.length;
    const currentPosition = state.current ? state.history.length + 1 : 0;

    return {
        queue: state.queue,
        current: state.current,
        history: state.history,
        playerStatus: state.playerStatus,
        totalQueued,
        currentPosition,
    };
}

// ---------------------------------------------------------------------------
// Hook for per-round status (optimized re-renders)
// ---------------------------------------------------------------------------

// Stable reference for idle state (prevents infinite loop with useSyncExternalStore)
const IDLE_STATE: CacheEntry = {
    url: null,
    status: 'idle',
    duration: null,
    currentTime: 0,
    errorMessage: undefined,
};

export function useRoundAudioStatus(
    debateId: string | undefined,
    roundIndex: number | undefined,
): RoundAudioState {
    const key = debateId != null && roundIndex != null ? getCacheKey(debateId, roundIndex) : null;

    const getStatusSnapshot = useCallback(() => {
        if (!key) return IDLE_STATE;
        // Return the actual cache entry (stable reference) or IDLE_STATE
        return cache.get(key) ?? IDLE_STATE;
    }, [key]);

    const state = useSyncExternalStore(subscribe, getStatusSnapshot, getStatusSnapshot);

    return {
        status: state.status,
        duration: state.duration,
        currentTime: state.currentTime,
        errorMessage: state.errorMessage,
    };
}

// ---------------------------------------------------------------------------
// Hook to check if a round is currently playing
// ---------------------------------------------------------------------------

export function useIsRoundPlaying(debateId: string, roundIndex: number): boolean {
    const queueState = useSyncExternalStore(subscribe, getQueueState, () => IDLE_QUEUE_STATE);
    return (
        queueState.current?.debateId === debateId &&
        queueState.current?.roundIndex === roundIndex &&
        queueState.playerStatus === 'playing'
    );
}

export function useIsRoundActive(debateId: string, roundIndex: number): boolean {
    const queueState = useSyncExternalStore(subscribe, getQueueState, () => IDLE_QUEUE_STATE);
    return (
        queueState.current?.debateId === debateId &&
        queueState.current?.roundIndex === roundIndex &&
        queueState.playerStatus !== 'idle'
    );
}

// ---------------------------------------------------------------------------
// Cleanup helper (call when debate changes to free memory)
// ---------------------------------------------------------------------------

export function clearDebateAudioCache(debateId: string): void {
    const keysToDelete: string[] = [];
    cache.forEach((entry, key) => {
        if (key.startsWith(`${debateId}:`)) {
            if (entry.url) {
                URL.revokeObjectURL(entry.url);
            }
            keysToDelete.push(key);
        }
    });
    keysToDelete.forEach((k) => cache.delete(k));
    notifyListeners();
}
