import { useNavigate, useRouteError } from 'react-router-dom';
import { Button } from '@/components/ui/Button';

export function EditErrorBoundary() {
    const error = useRouteError();
    const navigate = useNavigate();
    const message =
        error instanceof Error ? error.message : 'Something went wrong loading this debate.';

    return (
        <div className='flex min-h-full flex-col items-center justify-center gap-4 px-6 text-center'>
            <div className='rounded-2xl border border-agent-2/30 bg-agent-2-bg/60 px-6 py-5'>
                <p className='font-mono text-[10px] uppercase tracking-[0.18em] text-agent-2'>
                    Edit debate failed to load
                </p>
                <p className='mt-3 max-w-md text-sm leading-6 text-text-1'>{message}</p>
            </div>

            <Button type='button' variant='accent' onClick={() => navigate(0)} className='text-sm'>
                Try again
            </Button>
        </div>
    );
}
