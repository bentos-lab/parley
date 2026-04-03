export class ApiError extends Error {
    constructor(
        public readonly status: number,
        message: string,
        public readonly body?: unknown,
    ) {
        super(message);
        this.name = 'ApiError';
    }
}

// API_PREFIX is prepended to every backend path before fetching.
export const API_PREFIX = '/api';

export const BASE = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? '';

async function fetchWithMockRetry(
    path: string,
    init?: RequestInit,
    attempt = 0,
): Promise<Response> {
    const response = await fetch(`${BASE}${path}`, {
        headers: { 'Content-Type': 'application/json', ...init?.headers },
        ...init,
    });

    const contentType = response.headers.get('content-type') ?? '';
    const shouldRetryForMockStartup =
        import.meta.env.DEV && contentType.includes('text/html') && attempt < 2;

    if (!shouldRetryForMockStartup) {
        return response;
    }

    await new Promise((resolve) => setTimeout(resolve, 150 * (attempt + 1)));
    return fetchWithMockRetry(path, init, attempt + 1);
}

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
    const res = await fetchWithMockRetry(path, init);
    if (!res.ok) {
        const text = await res.text().catch(() => res.statusText);
        let parsedBody: unknown;
        try {
            parsedBody = text ? JSON.parse(text) : undefined;
        } catch {
            parsedBody = undefined;
        }
        throw new ApiError(res.status, text, parsedBody);
    }
    return res.json() as Promise<T>;
}

export async function requestVoid(path: string, init?: RequestInit): Promise<void> {
    const res = await fetchWithMockRetry(path, init);
    if (!res.ok) {
        const text = await res.text().catch(() => res.statusText);
        let parsedBody: unknown;
        try {
            parsedBody = text ? JSON.parse(text) : undefined;
        } catch {
            parsedBody = undefined;
        }
        throw new ApiError(res.status, text, parsedBody);
    }
}

export async function requestBlob(path: string, init?: RequestInit): Promise<Blob> {
    const res = await fetchWithMockRetry(path, init);
    if (!res.ok) {
        const text = await res.text().catch(() => res.statusText);
        let parsedBody: unknown;
        try {
            parsedBody = text ? JSON.parse(text) : undefined;
        } catch {
            parsedBody = undefined;
        }
        throw new ApiError(res.status, text, parsedBody);
    }
    return res.blob();
}
