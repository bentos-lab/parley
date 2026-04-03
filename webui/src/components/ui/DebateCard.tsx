import { useState } from 'react';
import { useNavigate, useRevalidator } from 'react-router-dom';
import { deleteDebate } from '@/services/api/debates';
import type { DebateSummary } from '@/types';

export interface DebateCardProps {
    debate: DebateSummary;
}

export function DebateCard({ debate }: DebateCardProps) {
    const navigate = useNavigate();
    const revalidator = useRevalidator();
    const [deleting, setDeleting] = useState(false);

    async function handleDelete(e: React.MouseEvent) {
        e.stopPropagation();
        if (deleting) return;
        setDeleting(true);
        try {
            await deleteDebate(debate.id);
            revalidator.revalidate();
        } catch {
            setDeleting(false);
        }
    }

    return (
        <div className='relative group rounded-xl border border-border bg-bg-panel transition-[border-color,background] hover:border-border-hi hover:bg-bg-surface'>
            <button
                type='button'
                onClick={() => navigate(`/debates/${encodeURIComponent(debate.id)}`)}
                className='w-full text-left p-5 cursor-pointer'
            >
                <div className='text-xl mb-2.5'>💬</div>
                <h3 className='text-[13px] font-semibold text-text-1 mb-1 truncate'>
                    {debate.name}
                </h3>
                <p className='text-xs text-text-3 line-clamp-2 leading-[1.5]'>{debate.topic}</p>
            </button>

            {/* Delete button — visible on hover */}
            <button
                type='button'
                onClick={handleDelete}
                disabled={deleting}
                aria-label='Delete debate'
                className='absolute top-2.5 right-2.5 flex h-6 w-6 items-center justify-center rounded-md text-text-3 opacity-0 transition-opacity group-hover:opacity-100 hover:bg-bg-elevated hover:text-red-500 disabled:cursor-not-allowed disabled:opacity-40'
            >
                {deleting ? (
                    <span className='h-3 w-3 animate-spin rounded-full border border-text-3 border-t-transparent' />
                ) : (
                    <svg width='12' height='12' viewBox='0 0 12 12' fill='none' aria-hidden='true'>
                        <path
                            d='M1 1l10 10M11 1L1 11'
                            stroke='currentColor'
                            strokeWidth='1.5'
                            strokeLinecap='round'
                        />
                    </svg>
                )}
            </button>
        </div>
    );
}
