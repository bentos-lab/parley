import { useRevalidator } from 'react-router-dom';
import { Button } from '@/components/ui/Button';

export interface ErrorStateProps {
    error?: Error | string;
}

export function ErrorState({ error }: ErrorStateProps) {
    const revalidator = useRevalidator();

    const errorMessage =
        error instanceof Error ? error.message : String(error || 'Failed to load debates');

    return (
        <div className='flex flex-col items-center justify-center py-16 px-4 text-center'>
            {/* Error icon placeholder */}
            <div className='w-20 h-20 rounded-full bg-agent-2-bg border-2 border-agent-2 flex items-center justify-center mb-6'>
                <span className='text-3xl text-agent-2'>⚠</span>
            </div>

            <h2 className='text-lg font-display font-medium text-text-1 mb-2'>
                Could not load debates
            </h2>
            <p className='text-text-2 text-sm mb-4 max-w-xs'>{errorMessage}</p>

            <Button variant='accent' onClick={() => revalidator.revalidate()}>
                Try Again
            </Button>
        </div>
    );
}
