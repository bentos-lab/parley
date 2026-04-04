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

            <div className='flex w-full max-w-[880px] items-center justify-between gap-4 rounded-xl border border-border bg-bg-surface px-5 py-4'>
                <div>
                    <p className='text-[11px] font-semibold uppercase tracking-[0.08em] text-text-3'>
                        Integration
                    </p>
                    <p className='mt-1 text-sm text-text-2'>
                        Pair WhatsApp to run `/parley` from your phone and keep debates within
                        reach.
                    </p>
                </div>
                <button
                    type='button'
                    onClick={() => navigate('/integrations/whatsapp')}
                    className='shrink-0 rounded-md border border-border-mid bg-bg-base px-3 py-2 text-xs font-semibold text-text-1 transition-colors hover:border-border-hi hover:bg-bg-hover cursor-pointer'
                >
                    Open WhatsApp
                </button>
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
