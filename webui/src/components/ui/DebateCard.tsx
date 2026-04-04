import { useState } from 'react';
import { useNavigate, useRevalidator } from 'react-router-dom';
import { DeleteConfirmModal } from './DeleteConfirmModal';
import { useToast } from './ToastProvider';
import { deleteDebate } from '@/services/api/debates';
import { getErrorMessage } from '@/services/api/http';
import type { DebateSummary } from '@/types';

export interface DebateCardProps {
    debate: DebateSummary;
}

export function DebateCard({ debate }: DebateCardProps) {
    const navigate = useNavigate();
    const revalidator = useRevalidator();
    const toast = useToast();
    const [deleting, setDeleting] = useState(false);
    const [confirmOpen, setConfirmOpen] = useState(false);

    async function handleDelete() {
        if (deleting) return;
        setDeleting(true);
        try {
            await deleteDebate(debate.id);
            revalidator.revalidate();
            setConfirmOpen(false);
        } catch (error) {
            toast.error(getErrorMessage(error, 'Failed to delete this debate.'), {
                title: 'Delete failed',
            });
        } finally {
            setDeleting(false);
        }
    }

    return (
        <>
            <div className='relative group rounded-xl border border-border bg-bg-panel transition-[border-color,background] hover:border-border-hi hover:bg-bg-surface'>
                <button
                    type='button'
                    onClick={() => navigate(`/debates/${encodeURIComponent(debate.id)}`)}
                    className='w-full cursor-pointer p-5 text-left'
                >
                    <div className='mb-2.5 text-xl'>💬</div>
                    <h3 className='mb-1 truncate text-[13px] font-semibold text-text-1'>
                        {debate.name}
                    </h3>
                    <p className='line-clamp-2 text-xs leading-normal text-text-3'>
                        {debate.topic}
                    </p>
                </button>

                <button
                    type='button'
                    onClick={(e) => {
                        e.stopPropagation();
                        setConfirmOpen(true);
                    }}
                    disabled={deleting}
                    aria-label='Delete debate'
                    className='absolute top-2.5 right-2.5 flex h-6 w-6 items-center justify-center rounded-md text-text-3 opacity-0 transition-opacity group-hover:opacity-100 hover:bg-bg-elevated hover:text-red-500 disabled:cursor-not-allowed disabled:opacity-40'
                >
                    {deleting ? (
                        <span className='h-3 w-3 animate-spin rounded-full border border-text-3 border-t-transparent' />
                    ) : (
                        <svg
                            width='11'
                            height='11'
                            viewBox='0 0 11 11'
                            fill='none'
                            aria-hidden='true'
                        >
                            <path
                                d='M1 2.5h9M3.5 2.5V1.75A.25.25 0 0 1 3.75 1.5h3.5a.25.25 0 0 1 .25.25V2.5M4 4.5v3M7 4.5v3M2 2.5l.75 6.75a.25.25 0 0 0 .25.25h5a.25.25 0 0 0 .25-.25L9 2.5'
                                stroke='currentColor'
                                strokeWidth='1'
                                strokeLinecap='round'
                                strokeLinejoin='round'
                            />
                        </svg>
                    )}
                </button>
            </div>

            {confirmOpen && (
                <DeleteConfirmModal
                    debate={debate}
                    onConfirm={handleDelete}
                    onCancel={() => !deleting && setConfirmOpen(false)}
                    deleting={deleting}
                />
            )}
        </>
    );
}
