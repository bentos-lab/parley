const STORAGE_KEY = 'parley:session-config';

interface StoredSessionConfig {
    turnCount: number;
}

type StoredConfigs = Record<string, StoredSessionConfig>;

/** Default number of turns when none is configured */
export const DEFAULT_TURN_COUNT = 6;

function readStore(): StoredConfigs {
    if (typeof window === 'undefined') {
        return {};
    }

    try {
        const raw = window.localStorage.getItem(STORAGE_KEY);
        if (!raw) {
            return {};
        }

        const parsed = JSON.parse(raw) as StoredConfigs;
        return typeof parsed === 'object' && parsed ? parsed : {};
    } catch {
        return {};
    }
}

export function writeStoredSessionConfig(debateId: string, config: { turnCount: number }) {
    if (typeof window === 'undefined') {
        return;
    }

    window.localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify({ ...readStore(), [debateId]: config }),
    );
}

export function readStoredSessionConfig(debateId: string): StoredSessionConfig | null {
    const value = readStore()[debateId];
    if (value && typeof value.turnCount === 'number') {
        return value;
    }
    return null;
}
