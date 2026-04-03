import type { DebateSummary } from '@/types';

export interface StatsBarProps {
    debates: DebateSummary[];
}

export function StatsBar({ debates }: StatsBarProps) {
    return (
        <div className='flex gap-6'>
            <div className='text-center'>
                <div className='font-display text-[28px] text-accent'>{debates.length}</div>
                <div className='text-[11px] text-text-3'>debates</div>
            </div>
        </div>
    );
}
