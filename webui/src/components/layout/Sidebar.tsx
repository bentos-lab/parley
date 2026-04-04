import { useState } from 'react';
import { NavLink, useLocation, useNavigate, useRevalidator } from 'react-router-dom';
import { DeleteConfirmModal } from '../ui/DeleteConfirmModal';
import { useToast } from '../ui/ToastProvider';
import { deleteDebate } from '@/services/api/debates';
import { getErrorMessage } from '@/services/api/http';
import type { DebateSummary } from '@/types';

export interface SidebarProps {
    debates: DebateSummary[];
    sidebarError?: string | null;
    collapsed?: boolean;
    onToggle?: () => void;
}

export function Sidebar({ debates, sidebarError, collapsed = false, onToggle }: SidebarProps) {
    const location = useLocation();
    const navigate = useNavigate();
    const revalidator = useRevalidator();
    const toast = useToast();

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
            setPendingDelete(null);
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
            <aside
                className={`hidden md:flex h-full shrink-0 flex-col border-r border-border bg-bg-panel transition-[width] duration-200 ease-out ${
                    collapsed ? 'w-14' : 'w-57'
                }`}
            >
                {/* Logo + collapse toggle */}
                <div
                    className={`flex h-14 items-center border-b border-border ${
                        collapsed ? 'justify-center px-2' : 'gap-2.5 px-4'
                    }`}
                >
                    <img src='/logo.png' alt='Parley' className='h-7 w-7 rounded-md object-cover' />
                    {!collapsed && (
                        <>
                            <span className='font-display text-[16px] text-text-1 tracking-[0.01em]'>
                                parley
                            </span>
                            <span className='font-mono text-[9px] text-text-3'>v0.1</span>
                        </>
                    )}
                    {/* Collapse toggle button */}
                    <button
                        type='button'
                        aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
                        title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
                        onClick={onToggle}
                        className={`flex h-7 w-7 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-hover hover:text-text-1 cursor-pointer ${
                            collapsed ? '' : 'ml-auto'
                        }`}
                    >
                        <svg
                            className={`h-3.5 w-3.5 transition-transform duration-200 ${collapsed ? 'rotate-180' : ''}`}
                            viewBox='0 0 16 16'
                            fill='none'
                            aria-hidden='true'
                        >
                            <path
                                d='M10 12L6 8l4-4'
                                stroke='currentColor'
                                strokeWidth='1.5'
                                strokeLinecap='round'
                                strokeLinejoin='round'
                            />
                        </svg>
                    </button>
                </div>

                {/* Workspace nav */}
                {!collapsed && (
                    <div className='px-4 pt-3.5 pb-1.5 text-[9px] font-semibold uppercase tracking-[0.1em] text-text-3'>
                        Workspace
                    </div>
                )}
                <nav className={collapsed ? 'px-1 pt-2' : 'px-1.5'}>
                    <NavLink
                        end
                        to='/debates'
                        title={collapsed ? 'Home' : undefined}
                        className={({ isActive }) =>
                            [
                                'flex items-center rounded-[5px] transition-colors cursor-pointer',
                                collapsed
                                    ? 'justify-center w-10 h-10 mx-auto my-1'
                                    : 'gap-2 px-3.5 py-[7px] mx-0 text-xs',
                                isActive && !activeDebateId
                                    ? 'bg-bg-elevated text-text-1'
                                    : 'text-text-2 hover:bg-bg-hover hover:text-text-1',
                            ].join(' ')
                        }
                    >
                        <svg
                            className={`shrink-0 ${collapsed ? 'w-4.5 h-4.5' : 'w-3.5 h-3.5 opacity-60'}`}
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
                        {!collapsed && 'Home'}
                    </NavLink>
                    <NavLink
                        to='/debates/new'
                        title={collapsed ? 'New debate' : undefined}
                        className={({ isActive }) =>
                            [
                                'flex items-center rounded-[5px] transition-colors cursor-pointer',
                                collapsed
                                    ? 'justify-center w-10 h-10 mx-auto my-1'
                                    : 'gap-2 px-3.5 py-[7px] mx-0 text-xs',
                                isActive
                                    ? 'bg-bg-elevated text-text-1'
                                    : 'text-text-2 hover:bg-bg-hover hover:text-text-1',
                            ].join(' ')
                        }
                    >
                        <svg
                            className={`shrink-0 ${collapsed ? 'w-4.5 h-4.5' : 'w-3.5 h-3.5 opacity-60'}`}
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
                        {!collapsed && 'New debate'}
                    </NavLink>
                    <NavLink
                        to='/settings'
                        title={collapsed ? 'Settings' : undefined}
                        className={({ isActive }) =>
                            [
                                'flex items-center rounded-[5px] transition-colors cursor-pointer',
                                collapsed
                                    ? 'justify-center w-10 h-10 mx-auto my-1'
                                    : 'gap-2 px-3.5 py-[7px] mx-0 text-xs',
                                isActive
                                    ? 'bg-bg-elevated text-text-1'
                                    : 'text-text-2 hover:bg-bg-hover hover:text-text-1',
                            ].join(' ')
                        }
                    >
                        <svg
                            className={`shrink-0 ${collapsed ? 'w-4.5 h-4.5' : 'w-3.5 h-3.5 opacity-60'}`}
                            viewBox='0 0 16 16'
                            fill='none'
                            aria-hidden='true'
                        >
                            <circle cx='8' cy='8' r='2' stroke='currentColor' strokeWidth='1.2' />
                            <path
                                d='M8 1v2M8 13v2M1 8h2M13 8h2M2.93 2.93l1.41 1.41M11.66 11.66l1.41 1.41M2.93 13.07l1.41-1.41M11.66 4.34l1.41-1.41'
                                stroke='currentColor'
                                strokeWidth='1.2'
                                strokeLinecap='round'
                            />
                        </svg>
                        {!collapsed && 'Settings'}
                    </NavLink>
                    <NavLink
                        to='/integrations/whatsapp'
                        title={collapsed ? 'WhatsApp' : undefined}
                        className={({ isActive }) =>
                            [
                                'flex items-center rounded-[5px] transition-colors cursor-pointer',
                                collapsed
                                    ? 'justify-center w-10 h-10 mx-auto my-1'
                                    : 'gap-2 px-3.5 py-[7px] mx-0 text-xs',
                                isActive
                                    ? 'bg-bg-elevated text-text-1'
                                    : 'text-text-2 hover:bg-bg-hover hover:text-text-1',
                            ].join(' ')
                        }
                    >
                        <svg
                            className={`shrink-0 ${collapsed ? 'w-4.5 h-4.5' : 'w-3.5 h-3.5 opacity-60'}`}
                            viewBox='0 0 24 24'
                            fill='none'
                            aria-hidden='true'
                        >
                            <path
                                d='M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347'
                                stroke='currentColor'
                                strokeWidth='1.2'
                                strokeLinecap='round'
                                strokeLinejoin='round'
                            />
                            <path
                                d='M12.045 21.794h-.005a11.882 11.882 0 0 1-5.683-1.448L.057 24l1.688-6.163a11.861 11.861 0 0 1-1.588-5.945C.16 5.335 5.495 0 12.05 0a11.821 11.821 0 0 1 8.413 3.488 11.821 11.821 0 0 1 3.48 8.413c-.003 6.558-5.339 11.893-11.893 11.893Z'
                                stroke='currentColor'
                                strokeWidth='1.2'
                                strokeLinejoin='round'
                            />
                        </svg>
                        {!collapsed && 'WhatsApp'}
                    </NavLink>
                </nav>

                {/* Saved debates — hidden when collapsed */}
                {!collapsed && (
                    <>
                        <div className='px-4 pt-3.5 pb-1.5 text-[9px] font-semibold uppercase tracking-[0.1em] text-text-3'>
                            Saved debates
                        </div>
                        <div className='flex-1 overflow-y-auto px-1.5 pb-2 scrollbar-thin'>
                            {sidebarError ? (
                                <div className='mx-2 rounded-lg border border-error/25 bg-error/5 px-3 py-2.5'>
                                    <p className='text-[10px] font-medium uppercase tracking-wider text-error'>
                                        Offline
                                    </p>
                                    <p className='mt-1 text-[11px] leading-4 text-text-3'>
                                        {sidebarError}
                                    </p>
                                </div>
                            ) : debates.length === 0 ? (
                                <p className='px-3.5 py-2 text-xs text-text-3'>No debates yet</p>
                            ) : null}
                            {debates.map((debate) => {
                                const isActive = activeDebateId === encodeURIComponent(debate.id);
                                return (
                                    <div key={debate.id} className='group relative my-px'>
                                        <button
                                            type='button'
                                            onClick={() =>
                                                navigate(
                                                    `/debates/${encodeURIComponent(debate.id)}`,
                                                )
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
                    </>
                )}

                {/* Spacer when collapsed */}
                {collapsed && <div className='flex-1' />}

                {/* Bottom section: user */}
                <div
                    className={`border-t border-border flex items-center ${
                        collapsed ? 'justify-center py-3 px-0' : 'gap-2 px-4 py-3'
                    }`}
                >
                    <div
                        className='flex h-[26px] w-[26px] shrink-0 items-center justify-center rounded-full bg-bg-elevated border border-border-mid text-[10px] text-text-2'
                        title={collapsed ? 'parley' : undefined}
                    >
                        P
                    </div>
                    {!collapsed && <span className='text-xs text-text-2'>parley</span>}
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
