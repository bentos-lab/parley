import { useRouteError } from 'react-router-dom';
import { ErrorState } from '@/components/browse/ErrorState';

export function ListErrorBoundary() {
    const error = useRouteError();

    return (
        <div style={{ padding: '1rem' }}>
            <ErrorState error={error instanceof Error ? error : undefined} />
        </div>
    );
}
