import { useRouteError } from 'react-router-dom';

export function AudioErrorBoundary() {
    const error = useRouteError();
    const message = error instanceof Error ? error.message : String(error);

    return (
        <div style={{ padding: '1rem', color: '#b05a3c' }}>
            <strong>Audio studio failed to load:</strong> {message}
        </div>
    );
}
