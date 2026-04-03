import { useRouteError } from 'react-router-dom';

export function RootErrorBoundary() {
    const error = useRouteError();
    const message = error instanceof Error ? error.message : 'An unexpected error occurred.';

    return (
        <div style={{ padding: '2rem', fontFamily: 'monospace' }}>
            <h1>Something went wrong</h1>
            <p style={{ marginTop: '1rem', opacity: 0.7 }}>{message}</p>
        </div>
    );
}
