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

interface ParsedErrorDetails {
    code?: string;
    message?: string;
}

function normalizeCode(value: unknown): string | undefined {
    if (typeof value === 'number' && Number.isFinite(value)) {
        return String(value);
    }

    if (typeof value === 'string') {
        const trimmed = value.trim();
        return trimmed ? trimmed : undefined;
    }

    return undefined;
}

function tryParseJson(value: string): unknown | null {
    try {
        return JSON.parse(value) as unknown;
    } catch {
        return null;
    }
}

function parseEmbeddedJsonError(value: string): ParsedErrorDetails | null {
    const trimmed = value.trim();
    const objectStart = trimmed.indexOf('{');
    const arrayStart = trimmed.indexOf('[');
    const candidates = [objectStart, arrayStart].filter((index) => index >= 0);
    const jsonStart = candidates.length > 0 ? Math.min(...candidates) : -1;

    if (jsonStart === -1) {
        return null;
    }

    const parsed = tryParseJson(trimmed.slice(jsonStart));
    if (!parsed) {
        return null;
    }

    const details = parseErrorDetails(parsed);
    if (!details.code && !details.message) {
        return null;
    }

    const prefix = trimmed.slice(0, jsonStart);
    const prefixCode = prefix.match(/\b(\d{3})\b/)?.[1];

    return {
        code: details.code ?? prefixCode,
        message: details.message,
    };
}

function parseErrorDetails(body: unknown): ParsedErrorDetails {
    if (!body) {
        return {};
    }

    if (typeof body === 'string') {
        const trimmed = body.trim();

        if (!trimmed) {
            return {};
        }

        const parsedJson = tryParseJson(trimmed);
        if (parsedJson) {
            return parseErrorDetails(parsedJson);
        }

        const embedded = parseEmbeddedJsonError(trimmed);
        if (embedded) {
            return embedded;
        }

        return { message: trimmed };
    }

    if (Array.isArray(body)) {
        for (const item of body) {
            const details = parseErrorDetails(item);
            if (details.code || details.message) {
                return details;
            }
        }

        return {};
    }

    if (typeof body !== 'object') {
        return {};
    }

    const record = body as Record<string, unknown>;
    const ownCode = normalizeCode(record.code) ?? normalizeCode(record.status);

    for (const key of ['message', 'detail', 'details']) {
        const value = record[key];
        if (typeof value === 'string' && value.trim()) {
            return {
                code: ownCode,
                message: value.trim(),
            };
        }
    }

    for (const key of ['error', 'errors']) {
        if (!(key in record)) {
            continue;
        }

        const nested = parseErrorDetails(record[key]);
        if (nested.code || nested.message) {
            return {
                code: nested.code ?? ownCode,
                message: nested.message,
            };
        }
    }

    for (const value of Object.values(record)) {
        if (typeof value === 'object' && value) {
            const nested = parseErrorDetails(value);
            if (nested.code || nested.message) {
                return {
                    code: nested.code ?? ownCode,
                    message: nested.message,
                };
            }
        }
    }

    return { code: ownCode };
}

function formatErrorMessage(
    code: string | undefined,
    message: string | undefined,
    fallback: string,
) {
    const safeMessage = message?.trim() || fallback;

    if (!code) {
        return safeMessage;
    }

    return `Error ${code}: ${safeMessage}`;
}

function normalizeResponseErrorMessage(status: number, text: string, statusText: string): string {
    const details = parseErrorDetails(text);

    if (details.code || details.message) {
        return formatErrorMessage(
            details.code ?? String(status),
            details.message,
            'Something went wrong',
        );
    }

    if (statusText.trim()) {
        return formatErrorMessage(String(status), statusText, 'Something went wrong');
    }

    return formatErrorMessage(String(status), undefined, 'Something went wrong');
}

export function getErrorMessage(error: unknown, fallback = 'Something went wrong.'): string {
    if (error instanceof ApiError) {
        const details = parseErrorDetails(error.body ?? error.message);
        return formatErrorMessage(details.code ?? String(error.status), details.message, fallback);
    }

    if (error instanceof Error) {
        const message = error.message.trim();

        if (!message) {
            return fallback;
        }

        if (/failed to fetch|networkerror|load failed/i.test(message)) {
            return 'Cannot reach the Parley backend. Check that the server is running and reachable.';
        }

        const details = parseErrorDetails(message);
        return formatErrorMessage(details.code, details.message, fallback);
    }

    return fallback;
}

async function buildApiError(response: Response): Promise<ApiError> {
    const text = await response.text().catch(() => response.statusText);
    let parsedBody: unknown;

    try {
        parsedBody = text ? JSON.parse(text) : undefined;
    } catch {
        parsedBody = undefined;
    }

    const details = parseErrorDetails(parsedBody);
    const message =
        details.code || details.message
            ? formatErrorMessage(
                  details.code ?? String(response.status),
                  details.message,
                  'Something went wrong',
              )
            : normalizeResponseErrorMessage(response.status, text, response.statusText);

    return new ApiError(response.status, message, parsedBody);
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
        throw await buildApiError(res);
    }
    return res.json() as Promise<T>;
}

export async function requestVoid(path: string, init?: RequestInit): Promise<void> {
    const res = await fetchWithMockRetry(path, init);
    if (!res.ok) {
        throw await buildApiError(res);
    }
}

export async function requestBlob(path: string, init?: RequestInit): Promise<Blob> {
    const res = await fetchWithMockRetry(path, init);
    if (!res.ok) {
        throw await buildApiError(res);
    }
    return res.blob();
}
