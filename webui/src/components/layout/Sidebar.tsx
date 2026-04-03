import { useState } from 'react';
import { NavLink, useLocation, useNavigate, useRevalidator } from 'react-router-dom';
import { deleteDebate } from '@/services/api/debates';
import type { DebateSummary } from '@/types';

export interface SidebarProps {
    debates: DebateSummary[];
}

function DeleteConfirmModal({
    debate,
    onConfirm,
    onCancel,
    deleting,
}: {
    debate: DebateSummary;
    onConfirm: () => void;
    onCancel: () => void;
    deleting: boolean;
}) {
    return (
        /* Backdrop */
        <div
            className='fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-[2px]'
            onClick={onCancel}
        >
            <div
                className='mx-4 w-full max-w-sm rounded-xl border border-border bg-bg-panel p-5 shadow-xl'
                onClick={(e) => e.stopPropagation()}
            >
                {/* Icon */}
                <div className='mb-3 flex h-9 w-9 items-center justify-center rounded-full bg-red-500/10'>
                    <svg width='16' height='16' viewBox='0 0 16 16' fill='none' aria-hidden='true'>
                        <path
                            d='M2 4h12M5 4V2.5A.5.5 0 0 1 5.5 2h5a.5.5 0 0 1 .5.5V4M6 7v4M10 7v4M3 4l1 9.5A.5.5 0 0 0 4.5 14h7a.5.5 0 0 0 .5-.5L13 4'
                            stroke='#ef4444'
                            strokeWidth='1.2'
                            strokeLinecap='round'
                            strokeLinejoin='round'
                        />
                    </svg>
                </div>

                <h2 className='mb-1 text-sm font-semibold text-text-1'>Delete debate?</h2>
                <p className='mb-4 text-[12px] text-text-3 leading-[1.6]'>
                    <span className='font-medium text-text-2'>{debate.name}</span> will be
                    permanently deleted from disk. This cannot be undone.
                </p>

                <div className='flex items-center justify-end gap-2'>
                    <button
                        type='button'
                        onClick={onCancel}
                        disabled={deleting}
                        className='rounded-md border border-border px-3.5 py-1.5 text-xs font-medium text-text-2 transition-colors hover:bg-bg-hover disabled:opacity-40 cursor-pointer'
                    >
                        Cancel
                    </button>
                    <button
                        type='button'
                        onClick={onConfirm}
                        disabled={deleting}
                        className='flex items-center gap-1.5 rounded-md bg-red-600 px-3.5 py-1.5 text-xs font-medium text-white transition-opacity hover:opacity-90 disabled:opacity-60 cursor-pointer'
                    >
                        {deleting && (
                            <span className='h-3 w-3 animate-spin rounded-full border border-white/40 border-t-white' />
                        )}
                        Delete
                    </button>
                </div>
            </div>
        </div>
    );
}

export function Sidebar({ debates }: SidebarProps) {
    const location = useLocation();
    const navigate = useNavigate();
    const revalidator = useRevalidator();

    const activeDebateId = location.pathname.match(/^\/debates\/([^/]+)/)?.[1];

    const [pendingDelete, setPendingDelete] = useState<DebateSummary | null>(null);
    const [deleting, setDeleting] = useState(false);

    async function handleConfirmDelete() {
        if (!pendingDelete || deleting) return;
        setDeleting(true);
        try {
            await deleteDebate(pendingDelete.id);
            // Navigate away if we were viewing the deleted debate
            if (activeDebateId === encodeURIComponent(pendingDelete.id)) {
                navigate('/debates');
            }
            revalidator.revalidate();
        } catch {
            // keep modal open on error — user can try again or cancel
        } finally {
            setDeleting(false);
            setPendingDelete(null);
        }
    }

    return (
        <>
            <aside className='hidden md:flex h-full w-57 shrink-0 flex-col border-r border-border bg-bg-panel'>
                {/* Logo */}
                <div className='flex h-14 items-center gap-2.5 border-b border-border px-4'>
                    <img src='/logo.png' alt='Parley' className='h-7 w-7 rounded-md object-cover' />
                    <span className='font-display text-[16px] text-text-1 tracking-[0.01em]'>
                        parley
                    </span>
                    <span className='ml-auto font-mono text-[9px] text-text-3'>v0.1</span>
                </div>

                {/* Workspace nav */}
                <div className='px-4 pt-3.5 pb-1.5 text-[9px] font-semibold uppercase tracking-[0.1em] text-text-3'>
                    Workspace
                </div>
                <nav className='px-1.5'>
                    <NavLink
                        end
                        to='/debates'
                        className={({ isActive }) =>
                            [
                                'flex items-center gap-2 rounded-[5px] px-3.5 py-[7px] mx-0 text-xs transition-colors cursor-pointer',
                                isActive && !activeDebateId
                                    ? 'bg-bg-elevated text-text-1'
                                    : 'text-text-2 hover:bg-bg-hover hover:text-text-1',
                            ].join(' ')
                        }
                    >
                        <svg
                            className='w-3.5 h-3.5 opacity-60 shrink-0'
                            viewBox='0 0 16 16'
                            fill='none'
                            aria-hidden='true'
                        >
                            <path
                                d='M2 6.5L8 2l6 4.5V14H2V6.5z'
                                stroke='currentColor'
                                strokeWidth='1.2'
                                strokeLinejoin='round'
                            />
                        </svg>
                        Home
                    </NavLink>
                    <NavLink
                        to='/debates/new'
                        className={({ isActive }) =>
                            [
                                'flex items-center gap-2 rounded-[5px] px-3.5 py-[7px] mx-0 text-xs transition-colors cursor-pointer',
                                isActive
                                    ? 'bg-bg-elevated text-text-1'
                                    : 'text-text-2 hover:bg-bg-hover hover:text-text-1',
                            ].join(' ')
                        }
                    >
                        <svg
                            className='w-3.5 h-3.5 opacity-60 shrink-0'
                            viewBox='0 0 16 16'
                            fill='none'
                            aria-hidden='true'
                        >
                            <circle cx='8' cy='8' r='6' stroke='currentColor' strokeWidth='1.2' />
                            <path
                                d='M8 5v6M5 8h6'
                                stroke='currentColor'
                                strokeWidth='1.2'
                                strokeLinecap='round'
                            />
                        </svg>
                        New debate
                    </NavLink>
                </nav>

                {/* Saved debates */}
                <div className='px-4 pt-3.5 pb-1.5 text-[9px] font-semibold uppercase tracking-[0.1em] text-text-3'>
                    Saved debates
                </div>
                <div className='flex-1 overflow-y-auto px-1.5 pb-2 scrollbar-thin'>
                    {debates.map((debate) => {
                        const isActive = activeDebateId === encodeURIComponent(debate.id);
                        return (
                            <div key={debate.id} className='group relative my-px'>
                                <button
                                    type='button'
                                    onClick={() =>
                                        navigate(`/debates/${encodeURIComponent(debate.id)}`)
                                    }
                                    className={[
                                        'w-full text-left rounded-[5px] px-3.5 py-2 pr-8 cursor-pointer transition-colors border',
                                        isActive
                                            ? 'bg-bg-elevated border-border-mid'
                                            : 'border-transparent hover:bg-bg-hover',
                                    ].join(' ')}
                                >
                                    <div className='text-xs text-text-1 truncate'>
                                        {debate.name}
                                    </div>
                                    <div className='text-[10px] text-text-3 truncate'>
                                        {debate.topic}
                                    </div>
                                </button>

                                {/* Delete button — visible on row hover */}
                                <button
                                    type='button'
                                    aria-label={`Delete "${debate.name}"`}
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        setPendingDelete(debate);
                                    }}
                                    className='absolute right-1 top-1/2 -translate-y-1/2 flex h-5 w-5 items-center justify-center rounded text-text-3 opacity-0 transition-opacity group-hover:opacity-100 hover:bg-bg-elevated hover:text-red-500 cursor-pointer'
                                >
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
                                </button>
                            </div>
                        );
                    })}
                </div>

                {/* Bottom user */}
                <div className='border-t border-border px-4 py-3 flex items-center gap-2'>
                    <div className='flex h-[26px] w-[26px] shrink-0 items-center justify-center rounded-full bg-bg-elevated border border-border-mid text-[10px] text-text-2'>
                        B
                    </div>
                    <span className='text-xs text-text-2'>parley</span>
                </div>
            </aside>

            {/* Delete confirmation modal (rendered outside aside for full-screen backdrop) */}
            {pendingDelete && (
                <DeleteConfirmModal
                    debate={pendingDelete}
                    onConfirm={handleConfirmDelete}
                    onCancel={() => !deleting && setPendingDelete(null)}
                    deleting={deleting}
                />
            )}
        </>
    );
}
