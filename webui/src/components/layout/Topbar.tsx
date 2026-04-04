import { useState } from 'react';
import { NavLink, useLocation, useNavigate, useParams } from 'react-router-dom';
import type { Debate } from '@/types';
import { Button } from '@/components/ui/Button';

export interface TopbarProps {
    debate?: Debate | null;
}

export function Topbar({ debate }: TopbarProps) {
    const [mobileNavOpen, setMobileNavOpen] = useState(false);
    const location = useLocation();
    const navigate = useNavigate();
    const params = useParams();

    const isDebateView = Boolean(params.debateId);
    const isCreateView = location.pathname === '/debates/new';
    const isEditView = location.pathname.endsWith('/edit');
    const isAudioView = location.pathname.endsWith('/audio');
    const detailPath = debate ? `/debates/${encodeURIComponent(debate.id)}` : null;

    let title = 'parley';
    let topic = 'Debate studio';
    if (isCreateView) {
        title = 'New debate';
        topic = 'POST /api/debates';
    } else if (debate) {
        title = debate.name;
        topic = debate.topic;
    }

    return (
        <>
            <header className='flex h-12 shrink-0 items-center gap-3 border-b border-border bg-bg-panel px-5'>
                {/* Mobile: brand + hamburger */}
                <div className='flex md:hidden items-center gap-2 flex-1 min-w-0'>
                    <span className='font-display text-[15px] text-text-1'>{title}</span>
                </div>

                {/* Desktop */}
                <span className='hidden md:block font-display text-[15px] text-text-1 shrink-0'>
                    {title}
                </span>
                <span className='hidden md:block flex-1 text-xs text-text-3 truncate ml-1.5 italic'>
                    {topic}
                </span>

                {/* Debate actions */}
                {isDebateView && debate && (
                    <div className='hidden md:flex items-center gap-1.5 shrink-0'>
                        {(isEditView || isAudioView) && detailPath ? (
                            <Button variant='ghost' onClick={() => navigate(detailPath)}>
                                Back to debate
                            </Button>
                        ) : null}
                        {!isEditView && !isAudioView ? (
                            <>
                                <Button
                                    variant='ghost'
                                    onClick={() =>
                                        navigate(`/debates/${encodeURIComponent(debate.id)}/audio`)
                                    }
                                >
                                    Audio studio
                                </Button>
                                <Button
                                    variant='ghost'
                                    onClick={() =>
                                        navigate(`/debates/${encodeURIComponent(debate.id)}/edit`)
                                    }
                                >
                                    Edit
                                </Button>
                            </>
                        ) : null}
                    </div>
                )}

                {/* Mobile hamburger */}
                <button
                    type='button'
                    aria-label={mobileNavOpen ? 'Close navigation' : 'Open navigation'}
                    aria-expanded={mobileNavOpen}
                    onClick={() => setMobileNavOpen((v) => !v)}
                    className='md:hidden flex items-center justify-center w-8 h-8 rounded text-text-2 hover:text-text-1 hover:bg-bg-hover transition-colors cursor-pointer'
                >
                    {mobileNavOpen ? (
                        <svg
                            viewBox='0 0 24 24'
                            fill='none'
                            stroke='currentColor'
                            strokeWidth='2'
                            strokeLinecap='round'
                            strokeLinejoin='round'
                            className='w-4 h-4'
                            aria-hidden='true'
                        >
                            <line x1='18' y1='6' x2='6' y2='18' />
                            <line x1='6' y1='6' x2='18' y2='18' />
                        </svg>
                    ) : (
                        <svg
                            viewBox='0 0 24 24'
                            fill='none'
                            stroke='currentColor'
                            strokeWidth='2'
                            strokeLinecap='round'
                            strokeLinejoin='round'
                            className='w-4 h-4'
                            aria-hidden='true'
                        >
                            <line x1='3' y1='8' x2='21' y2='8' />
                            <line x1='3' y1='16' x2='21' y2='16' />
                        </svg>
                    )}
                </button>
            </header>

            {/* Mobile nav drawer */}
            {mobileNavOpen && (
                <nav className='md:hidden border-b border-border bg-bg-panel px-4 py-3 space-y-1'>
                    <NavLink
                        end
                        to='/debates'
                        onClick={() => setMobileNavOpen(false)}
                        className={({ isActive }) =>
                            [
                                'flex items-center gap-2 rounded px-3 py-2 text-xs transition-colors',
                                isActive
                                    ? 'bg-bg-elevated text-text-1'
                                    : 'text-text-2 hover:bg-bg-hover hover:text-text-1',
                            ].join(' ')
                        }
                    >
                        Home
                    </NavLink>
                    <NavLink
                        to='/debates/new'
                        onClick={() => setMobileNavOpen(false)}
                        className={({ isActive }) =>
                            [
                                'flex items-center gap-2 rounded px-3 py-2 text-xs transition-colors',
                                isActive
                                    ? 'bg-bg-elevated text-text-1'
                                    : 'text-text-2 hover:bg-bg-hover hover:text-text-1',
                            ].join(' ')
                        }
                    >
                        New debate
                    </NavLink>
                    <NavLink
                        to='/settings'
                        onClick={() => setMobileNavOpen(false)}
                        className={({ isActive }) =>
                            [
                                'flex items-center gap-2 rounded px-3 py-2 text-xs transition-colors',
                                isActive
                                    ? 'bg-bg-elevated text-text-1'
                                    : 'text-text-2 hover:bg-bg-hover hover:text-text-1',
                            ].join(' ')
                        }
                    >
                        Settings
                    </NavLink>
                    {isDebateView && debate && (
                        <>
                            {(isEditView || isAudioView) && detailPath ? (
                                <button
                                    type='button'
                                    onClick={() => {
                                        setMobileNavOpen(false);
                                        navigate(detailPath);
                                    }}
                                    className='w-full text-left flex items-center gap-2 rounded px-3 py-2 text-xs text-text-2 hover:bg-bg-hover hover:text-text-1 transition-colors cursor-pointer'
                                >
                                    Back to debate
                                </button>
                            ) : null}
                            <button
                                type='button'
                                onClick={() => {
                                    setMobileNavOpen(false);
                                    navigate(`/debates/${encodeURIComponent(debate.id)}/audio`);
                                }}
                                className='w-full text-left flex items-center gap-2 rounded px-3 py-2 text-xs text-text-2 hover:bg-bg-hover hover:text-text-1 transition-colors cursor-pointer'
                            >
                                Audio studio
                            </button>
                            <button
                                type='button'
                                onClick={() => {
                                    setMobileNavOpen(false);
                                    navigate(`/debates/${encodeURIComponent(debate.id)}/edit`);
                                }}
                                className='w-full text-left flex items-center gap-2 rounded px-3 py-2 text-xs text-text-2 hover:bg-bg-hover hover:text-text-1 transition-colors cursor-pointer'
                            >
                                Edit
                            </button>
                        </>
                    )}
                </nav>
            )}
        </>
    );
}
