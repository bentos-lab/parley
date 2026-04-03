import type { LoaderFunctionArgs } from 'react-router-dom';
import { useLoaderData, useNavigate } from 'react-router-dom';
import { Suspense } from 'react';
import { listDebates } from '@/services/api/debates';
import { DebateSummaryListSchema } from '@/services/api/schemas';
import type { DebateSummary } from '@/types';
import { BrowseLayout } from '@/components/browse/Layout';
import { DebateCard } from '@/components/ui/DebateCard';
import { StatsBar } from '@/components/browse/StatsBar';
import { SkeletonGrid } from '@/components/browse/SkeletonCard';
import { EmptyState } from '@/components/browse/EmptyState';

// eslint-disable-next-line react-refresh/only-export-components
export async function loader(_args: LoaderFunctionArgs) {
    const data = await listDebates();
    return DebateSummaryListSchema.parse(data);
}

export function Component() {
    const debates = useLoaderData() as DebateSummary[];
    const navigate = useNavigate();

    return (
        <BrowseLayout>
            {/* Hero */}
            <div className='text-center'>
                <h1 className='font-display text-4xl text-text-1 mb-2'>Good morning.</h1>
                <p className='text-sm text-text-3 max-w-[420px] leading-[1.6]'>
                    Choose a debate to resume, or create a new one.
                </p>
            </div>

            {debates.length === 0 ? (
                <EmptyState />
            ) : (
                <Suspense fallback={<SkeletonGrid count={3} />}>
                    <div className='grid w-full max-w-[880px] grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3'>
                        {debates.map((debate) => (
                            <DebateCard key={debate.id} debate={debate} />
                        ))}
                        {/* New debate card */}
                        <button
                            type='button'
                            onClick={() => navigate('/debates/new')}
                            className='flex flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-border-mid bg-transparent p-5 text-text-3 cursor-pointer transition-colors hover:border-border-hi hover:bg-bg-surface hover:text-text-2'
                        >
                            <span className='text-2xl'>+</span>
                            <span className='text-xs font-semibold'>New debate</span>
                        </button>
                    </div>
                </Suspense>
            )}

            <StatsBar debates={debates} />
        </BrowseLayout>
    );
}

export { ListErrorBoundary as ErrorBoundary } from './ErrorBoundary';
