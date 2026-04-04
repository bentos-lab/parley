import { useCallback, useMemo, useRef, useState } from 'react';
import type { LoaderFunctionArgs } from 'react-router-dom';
import { useLoaderData } from 'react-router-dom';
import { useToast } from '@/components/ui/ToastProvider';
import { getDebate } from '@/services/api/debates';
import { getErrorMessage } from '@/services/api/http';
import { debateAudioUrl } from '@/services/api/audio';
import { DebateSchema } from '@/services/api/schemas';
import { useRoundAudioCache, useRoundAudioStatus } from '@/hooks/useRoundAudioCache';
import type { Debate } from '@/types';

// ---------------------------------------------------------------------------
// SVG Icons
// ---------------------------------------------------------------------------
function IconPlay({ className }: { className?: string }) {
    return (
        <svg viewBox='0 0 24 24' fill='currentColor' className={className ?? 'w-4 h-4'} aria-hidden>
            <path d='M6 4.75a.75.75 0 0 1 1.14-.64l12 7.25a.75.75 0 0 1 0 1.28l-12 7.25A.75.75 0 0 1 6 19.25v-14.5z' />
        </svg>
    );
}

function IconPause({ className }: { className?: string }) {
    return (
        <svg viewBox='0 0 24 24' fill='currentColor' className={className ?? 'w-4 h-4'} aria-hidden>
            <rect x='5' y='4' width='4' height='16' rx='1.25' />
            <rect x='15' y='4' width='4' height='16' rx='1.25' />
        </svg>
    );
}

function IconSkipBack15({ className }: { className?: string }) {
    return (
        <svg
            viewBox='0 0 24 24'
            fill='none'
            stroke='currentColor'
            strokeWidth='1.75'
            strokeLinecap='round'
            strokeLinejoin='round'
            className={className ?? 'w-[15px] h-[15px]'}
            aria-hidden
        >
            <path d='M3 12a9 9 0 1 0 .75-3.6' />
            <polyline points='3 4.5 3 8.5 7 8.5' />
            <text
                x='13.5'
                y='15'
                fontSize='5.5'
                textAnchor='middle'
                stroke='none'
                fill='currentColor'
                fontWeight='700'
                fontFamily='system-ui,sans-serif'
            >
                15
            </text>
        </svg>
    );
}

function IconSkipForward15({ className }: { className?: string }) {
    return (
        <svg
            viewBox='0 0 24 24'
            fill='none'
            stroke='currentColor'
            strokeWidth='1.75'
            strokeLinecap='round'
            strokeLinejoin='round'
            className={className ?? 'w-[15px] h-[15px]'}
            aria-hidden
        >
            <path d='M21 12a9 9 0 1 1-.75-3.6' />
            <polyline points='21 4.5 21 8.5 17 8.5' />
            <text
                x='10.5'
                y='15'
                fontSize='5.5'
                textAnchor='middle'
                stroke='none'
                fill='currentColor'
                fontWeight='700'
                fontFamily='system-ui,sans-serif'
            >
                15
            </text>
        </svg>
    );
}

function IconDownload({ className }: { className?: string }) {
    return (
        <svg
            viewBox='0 0 24 24'
            fill='none'
            stroke='currentColor'
            strokeWidth='1.75'
            strokeLinecap='round'
            strokeLinejoin='round'
            className={className ?? 'w-[13px] h-[13px]'}
            aria-hidden
        >
            <path d='M12 3v12M7.5 11l4.5 4.5 4.5-4.5' />
            <path d='M3 19h18' />
        </svg>
    );
}

function IconPlaySmall() {
    return (
        <svg viewBox='0 0 24 24' fill='currentColor' className='w-2.5 h-2.5' aria-hidden>
            <path d='M6 4.75a.75.75 0 0 1 1.14-.64l12 7.25a.75.75 0 0 1 0 1.28l-12 7.25A.75.75 0 0 1 6 19.25v-14.5z' />
        </svg>
    );
}

function IconPauseSmall() {
    return (
        <svg viewBox='0 0 24 24' fill='currentColor' className='w-2.5 h-2.5' aria-hidden>
            <rect x='5' y='4' width='4' height='16' rx='1.25' />
            <rect x='15' y='4' width='4' height='16' rx='1.25' />
        </svg>
    );
}

function formatTime(seconds: number): string {
    const m = Math.floor(seconds / 60);
    const s = Math.floor(seconds % 60);
    return `${m}:${String(s).padStart(2, '0')}`;
}

const AGENT_COLORS = ['#4e8f68', '#b05a3c', '#5278a8', '#8a62c0'];
const USER_COLOR = '#7a9060';

function agentColor(index: number) {
    return AGENT_COLORS[index % AGENT_COLORS.length];
}

function agentIndex(debate: Debate, agentId: string) {
    return debate.agents.findIndex((a) => a.id === agentId);
}

// eslint-disable-next-line react-refresh/only-export-components
export async function loader({ params }: LoaderFunctionArgs) {
    const data = await getDebate(params.debateId!);
    return DebateSchema.parse({ id: params.debateId, ...data });
}

// ---------------------------------------------------------------------------
// Waveform (decorative)
// ---------------------------------------------------------------------------
function Waveform({ progress, onSeek }: { progress: number; onSeek: (pct: number) => void }) {
    // eslint-disable-next-line react-hooks/purity
    const bars = useMemo(() => Array.from({ length: 80 }, () => 20 + Math.random() * 60), []);
    const containerRef = useRef<HTMLDivElement>(null);

    function handleClick(e: React.MouseEvent<HTMLDivElement>) {
        const rect = containerRef.current?.getBoundingClientRect();
        if (!rect) return;
        const pct = ((e.clientX - rect.left) / rect.width) * 100;
        onSeek(Math.max(0, Math.min(100, pct)));
    }

    return (
        <div
            ref={containerRef}
            className='relative h-[52px] rounded-[5px] border border-border bg-bg-surface overflow-hidden cursor-pointer mb-2.5'
            onClick={handleClick}
        >
            <div className='flex items-center gap-[1.5px] h-full px-2.5 py-1.5'>
                {bars.map((h, i) => (
                    <div
                        key={i}
                        className='flex-1 min-w-[2px] rounded-[1px] transition-colors'
                        style={{
                            height: `${h}%`,
                            background:
                                i < (bars.length * progress) / 100
                                    ? '#c9a44a'
                                    : 'var(--border-mid)',
                        }}
                    />
                ))}
            </div>
            <div
                className='absolute left-0 top-0 h-full pointer-events-none border-r border-accent'
                style={{ width: `${progress}%`, background: 'rgba(201,164,74,0.12)' }}
            />
        </div>
    );
}

// ---------------------------------------------------------------------------
// Chapters with per-turn audio play buttons (uses shared cache)
// ---------------------------------------------------------------------------

function IconSkipBack15Small() {
    return (
        <svg
            viewBox='0 0 24 24'
            fill='none'
            stroke='currentColor'
            strokeWidth='2'
            strokeLinecap='round'
            strokeLinejoin='round'
            className='w-2.5 h-2.5'
            aria-hidden
        >
            <path d='M3 12a9 9 0 1 0 .75-3.6' />
            <polyline points='3 4.5 3 8.5 7 8.5' />
        </svg>
    );
}

function IconSkipForward15Small() {
    return (
        <svg
            viewBox='0 0 24 24'
            fill='none'
            stroke='currentColor'
            strokeWidth='2'
            strokeLinecap='round'
            strokeLinejoin='round'
            className='w-2.5 h-2.5'
            aria-hidden
        >
            <path d='M21 12a9 9 0 1 1-.75-3.6' />
            <polyline points='21 4.5 21 8.5 17 8.5' />
        </svg>
    );
}

function ChapterRow({ debate, roundIndex }: { debate: Debate; roundIndex: number }) {
    const round = debate.rounds[roundIndex];
    const isUser = !round.agent_id;
    const agent = isUser ? null : debate.agents.find((a) => a.id === round.agent_id);
    const idx = isUser ? -1 : agentIndex(debate, round.agent_id);
    const color = isUser ? USER_COLOR : agentColor(idx);

    const audioState = useRoundAudioStatus(debate.id, roundIndex);
    const { play, togglePause, seekBy } = useRoundAudioCache();

    const isPlaying = audioState.status === 'playing';
    const isLoading = audioState.status === 'loading';
    const hasError = audioState.status === 'error';
    // isActive: round is loaded and ready or currently playing — enables seek buttons
    const isActive = isPlaying || audioState.status === 'ready';

    return (
        <div
            className={`flex items-center gap-2.5 py-[7px] group ${roundIndex < debate.rounds.length - 1 ? 'border-b border-border' : ''}`}
        >
            <span className='font-mono text-[9px] text-text-3 w-4'>
                {String(roundIndex + 1).padStart(2, '0')}
            </span>
            <span className='h-1.5 w-1.5 rounded-full shrink-0' style={{ background: color }} />
            <span className='flex-1 text-xs text-text-2 group-hover:text-text-1 transition-colors'>
                {agent ? agent.name : 'You'} — Round {roundIndex + 1}
            </span>
            <span className='font-mono text-[10px] text-text-3 mr-2'>
                {audioState.duration != null ? formatTime(audioState.duration) : '—'}
            </span>

            {/* Mini audio controls */}
            <div className='flex items-center gap-1'>
                {/* Skip back 15s */}
                <button
                    type='button'
                    onClick={() => seekBy(-15)}
                    className='flex h-5 w-5 shrink-0 items-center justify-center rounded-full border border-border text-text-3 cursor-pointer transition-colors hover:border-accent hover:text-accent disabled:opacity-30 disabled:cursor-not-allowed'
                    disabled={!isActive || isLoading}
                    aria-label='Skip back 15 seconds'
                    title='Skip back 15 seconds'
                >
                    <IconSkipBack15Small />
                </button>

                {/* Play/Pause toggle */}
                <button
                    type='button'
                    onClick={() => {
                        if (isPlaying) {
                            togglePause();
                        } else {
                            void play(debate.id, roundIndex);
                        }
                    }}
                    className='flex h-6 w-6 shrink-0 items-center justify-center rounded-full border border-border text-text-3 cursor-pointer transition-colors hover:border-accent hover:text-accent disabled:opacity-40 disabled:cursor-wait'
                    disabled={isLoading}
                    aria-label={
                        isPlaying ? `Pause round ${roundIndex + 1}` : `Play round ${roundIndex + 1}`
                    }
                    title={
                        hasError
                            ? 'Round audio unavailable'
                            : isLoading
                              ? 'Loading…'
                              : isPlaying
                                ? `Pause round ${roundIndex + 1}`
                                : `Play round ${roundIndex + 1}`
                    }
                >
                    {isLoading ? (
                        <svg
                            viewBox='0 0 24 24'
                            fill='none'
                            stroke='currentColor'
                            strokeWidth='2'
                            strokeLinecap='round'
                            className='w-2.5 h-2.5 animate-spin'
                            aria-hidden
                        >
                            <circle cx='12' cy='12' r='9' strokeOpacity='0.25' />
                            <path d='M12 3a9 9 0 0 1 9 9' />
                        </svg>
                    ) : hasError ? (
                        <svg
                            viewBox='0 0 24 24'
                            fill='none'
                            stroke='currentColor'
                            strokeWidth='2'
                            strokeLinecap='round'
                            className='w-2.5 h-2.5 text-error'
                            aria-hidden
                        >
                            <circle cx='12' cy='12' r='9' />
                            <line x1='12' y1='8' x2='12' y2='12' />
                            <circle cx='12' cy='16' r='0.75' fill='currentColor' />
                        </svg>
                    ) : isPlaying ? (
                        <IconPauseSmall />
                    ) : (
                        <IconPlaySmall />
                    )}
                </button>

                {/* Skip forward 15s */}
                <button
                    type='button'
                    onClick={() => seekBy(15)}
                    className='flex h-5 w-5 shrink-0 items-center justify-center rounded-full border border-border text-text-3 cursor-pointer transition-colors hover:border-accent hover:text-accent disabled:opacity-30 disabled:cursor-not-allowed'
                    disabled={!isActive || isLoading}
                    aria-label='Skip forward 15 seconds'
                    title='Skip forward 15 seconds'
                >
                    <IconSkipForward15Small />
                </button>
            </div>
        </div>
    );
}

function Chapters({ debate }: { debate: Debate }) {
    return (
        <div className='rounded-xl border border-border bg-bg-panel p-4 mt-4'>
            <div className='mb-2 text-[9px] font-semibold uppercase tracking-[0.08em] text-text-3'>
                Chapters
            </div>

            <div className='flex items-center gap-2.5 py-[7px] border-b border-border'>
                <span className='font-mono text-[9px] text-text-3 w-4'>00</span>
                <span
                    className='h-1.5 w-1.5 rounded-full shrink-0'
                    style={{ background: 'var(--text-3)' }}
                />
                <span className='flex-1 text-xs text-text-2'>Intro</span>
                <span className='font-mono text-[10px] text-text-3'>0:00</span>
            </div>

            {debate.rounds.map((_, i) => (
                <ChapterRow key={i} debate={debate} roundIndex={i} />
            ))}
        </div>
    );
}

// ---------------------------------------------------------------------------
// Audio info card (read-only, right column)
// ---------------------------------------------------------------------------
function AudioInfoCard({ debate }: { debate: Debate }) {
    return (
        <div className='rounded-xl border border-border bg-bg-panel p-4 space-y-3'>
            <div className='text-[9px] font-semibold uppercase tracking-[0.08em] text-text-3'>
                Audio info
            </div>
            <div className='space-y-[9px]'>
                <div className='flex justify-between items-center'>
                    <span className='text-[11px] text-text-3'>Format</span>
                    <span className='font-mono text-[10px] text-text-2'>.wav</span>
                </div>
                <div className='flex justify-between items-center'>
                    <span className='text-[11px] text-text-3'>TTS provider</span>
                    <span className='font-mono text-[10px] text-accent'>{debate.ttsProvider}</span>
                </div>
                <div className='flex justify-between items-center'>
                    <span className='text-[11px] text-text-3'>Rounds</span>
                    <span className='font-mono text-[10px] text-text-2'>
                        {debate.rounds.length}
                    </span>
                </div>
                <div className='flex justify-between items-center'>
                    <span className='text-[11px] text-text-3'>Agents</span>
                    <span className='font-mono text-[10px] text-text-2'>
                        {debate.agents.length}
                    </span>
                </div>
            </div>
        </div>
    );
}

// ---------------------------------------------------------------------------
// Agent profiles card (right column)
// ---------------------------------------------------------------------------
function AgentProfilesCard({ debate }: { debate: Debate }) {
    return (
        <div className='rounded-xl border border-border bg-bg-panel p-4 mt-4'>
            <div className='text-[9px] font-semibold uppercase tracking-[0.08em] text-text-3 mb-3'>
                Agents
            </div>
            <div className='space-y-3'>
                {debate.agents.map((agent, i) => {
                    const color = agentColor(i);
                    return (
                        <div
                            key={agent.id}
                            className='rounded-lg border border-border bg-bg-surface p-3 space-y-1.5'
                        >
                            <div className='flex items-center gap-2'>
                                <span
                                    className='h-2 w-2 rounded-full shrink-0'
                                    style={{ background: color }}
                                />
                                <span className='text-[13px] font-medium text-text-1 truncate'>
                                    {agent.name}
                                </span>
                            </div>
                            <div className='space-y-1 pt-0.5'>
                                {agent.stance && (
                                    <div className='flex items-start gap-1.5'>
                                        <span className='text-[10px] text-text-3 w-10 shrink-0'>
                                            Stance
                                        </span>
                                        <span className='text-[10px] text-text-2 leading-tight'>
                                            {agent.stance}
                                        </span>
                                    </div>
                                )}
                                {agent.voiceName && (
                                    <div className='flex items-center gap-1.5'>
                                        <span className='text-[10px] text-text-3 w-10 shrink-0'>
                                            Voice
                                        </span>
                                        <span className='font-mono text-[10px] text-accent'>
                                            {agent.voiceName}
                                        </span>
                                    </div>
                                )}
                            </div>
                        </div>
                    );
                })}
            </div>
        </div>
    );
}

// ---------------------------------------------------------------------------
// Main Audio Studio component
// ---------------------------------------------------------------------------
export function Component() {
    const debate = useLoaderData() as Debate;
    const audioRef = useRef<HTMLAudioElement>(null);
    const toast = useToast();

    const [isPlaying, setIsPlaying] = useState(false);
    const [progress, setProgress] = useState(0);
    const [speed, setSpeed] = useState('1×');
    const [audioError, setAudioError] = useState(false);
    const [audioDuration, setAudioDuration] = useState<number | null>(null);
    const [currentTime, setCurrentTime] = useState(0);
    const [audioLoading, setAudioLoading] = useState(true);
    const [downloading, setDownloading] = useState(false);

    const fullAudioUrl = debateAudioUrl(debate.id);

    async function handleDownload() {
        if (downloading) return;
        setDownloading(true);
        try {
            const res = await fetch(fullAudioUrl);
            if (!res.ok) throw new Error('fetch failed');
            const blob = await res.blob();
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${debate.id}.wav`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(url);
        } catch (error) {
            toast.error(getErrorMessage(error, 'Failed to download the debate audio.'), {
                title: 'Download failed',
            });
        } finally {
            setDownloading(false);
        }
    }

    const handleLoadedMetadata = useCallback(() => {
        const el = audioRef.current;
        if (el && !isNaN(el.duration) && isFinite(el.duration)) {
            setAudioDuration(el.duration);
            setAudioLoading(false);
        }
    }, []);

    const handleTimeUpdate = useCallback(() => {
        const el = audioRef.current;
        if (el && el.duration) {
            setCurrentTime(el.currentTime);
            setProgress((el.currentTime / el.duration) * 100);
        }
    }, []);

    function togglePlay() {
        const el = audioRef.current;
        if (!el) return;
        if (isPlaying) {
            el.pause();
        } else {
            void el.play().catch(() => setAudioError(true));
        }
        setIsPlaying(!isPlaying);
    }

    function cycleSpeed() {
        const speeds = ['0.75×', '1×', '1.25×', '1.5×', '2×'];
        const idx = speeds.indexOf(speed);
        const next = speeds[(idx + 1) % speeds.length];
        setSpeed(next);
        if (audioRef.current) audioRef.current.playbackRate = parseFloat(next.replace('×', ''));
    }

    // Display formatted times
    const durationDisplay = audioDuration != null ? formatTime(audioDuration) : '—:——';
    const currentTimeDisplay = formatTime(currentTime);

    return (
        <div className='flex flex-col flex-1 overflow-y-auto p-7 md:p-8 gap-5 scrollbar-thin'>
            <div className='mx-auto w-full max-w-5xl'>
                <div className='flex items-center justify-between gap-3 shrink-0 mb-5'>
                    <h1 className='font-display text-xl text-text-1'>Audio studio</h1>
                </div>

                <div className='grid grid-cols-1 lg:grid-cols-[1fr_280px] gap-5'>
                    {/* Left column: Player + Chapters */}
                    <div>
                        {/* Hidden real audio element */}
                        <audio
                            ref={audioRef}
                            src={fullAudioUrl}
                            preload='metadata'
                            onLoadedMetadata={handleLoadedMetadata}
                            onTimeUpdate={handleTimeUpdate}
                            onEnded={() => setIsPlaying(false)}
                            onError={() => setAudioError(true)}
                        />

                        {/* Player card */}
                        <div className='rounded-xl border border-border bg-bg-panel p-5'>
                            <div className='flex gap-3.5 items-center mb-[18px]'>
                                <div className='flex h-16 w-16 shrink-0 items-center justify-center rounded-lg bg-accent-bg border border-accent-dim font-display text-xl text-accent'>
                                    弁
                                </div>
                                <div className='flex-1 min-w-0'>
                                    <div className='font-display text-[15px] text-text-1 mb-[3px] truncate'>
                                        {debate.name}
                                    </div>
                                    <div className='text-[11px] text-text-3'>
                                        {audioDuration != null
                                            ? formatTime(audioDuration)
                                            : audioLoading
                                              ? 'Loading...'
                                              : '—'}{' '}
                                        · {debate.rounds.length} rounds · {debate.agents.length}{' '}
                                        agents
                                    </div>
                                </div>
                            </div>

                            {audioError ? (
                                <div className='mb-4 rounded-lg border border-error/30 bg-error/10 px-3 py-2 text-xs text-error'>
                                    Audio unavailable — TTS may not be configured for this debate.
                                </div>
                            ) : null}

                            <Waveform
                                progress={progress}
                                onSeek={(pct) => {
                                    setProgress(pct);
                                    if (audioRef.current && audioRef.current.duration) {
                                        audioRef.current.currentTime =
                                            (pct / 100) * audioRef.current.duration;
                                    }
                                }}
                            />

                            <div className='flex justify-between font-mono text-[10px] text-text-3 mb-3.5'>
                                <span>{currentTimeDisplay}</span>
                                <span>{durationDisplay}</span>
                            </div>

                            <div className='flex items-center justify-center gap-3.5'>
                                <button
                                    type='button'
                                    className='flex h-[30px] w-[30px] items-center justify-center rounded-full border border-border-mid text-text-2 cursor-pointer transition-colors hover:bg-bg-hover hover:text-text-1'
                                    title='Skip back 15 seconds'
                                    aria-label='Skip back 15 seconds'
                                    onClick={() => {
                                        if (audioRef.current)
                                            audioRef.current.currentTime = Math.max(
                                                0,
                                                audioRef.current.currentTime - 15,
                                            );
                                    }}
                                >
                                    <IconSkipBack15 />
                                </button>
                                <button
                                    type='button'
                                    className='flex h-11 w-11 items-center justify-center rounded-full border border-border-hi text-text-2 cursor-pointer transition-colors hover:bg-bg-hover hover:text-text-1'
                                    title={isPlaying ? 'Pause' : 'Play'}
                                    aria-label={isPlaying ? 'Pause' : 'Play'}
                                    onClick={togglePlay}
                                >
                                    {isPlaying ? (
                                        <IconPause className='w-[18px] h-[18px]' />
                                    ) : (
                                        <IconPlay className='w-[18px] h-[18px]' />
                                    )}
                                </button>
                                <button
                                    type='button'
                                    className='flex h-[30px] w-[30px] items-center justify-center rounded-full border border-border-mid text-text-2 cursor-pointer transition-colors hover:bg-bg-hover hover:text-text-1'
                                    title='Skip forward 15 seconds'
                                    aria-label='Skip forward 15 seconds'
                                    onClick={() => {
                                        if (audioRef.current) audioRef.current.currentTime += 15;
                                    }}
                                >
                                    <IconSkipForward15 />
                                </button>
                                <button
                                    type='button'
                                    className='rounded border border-border bg-bg-surface px-[7px] py-0.5 font-mono text-[11px] text-text-2 cursor-pointer hover:text-text-1 transition-colors'
                                    title='Change playback speed'
                                    aria-label={`Playback speed: ${speed}`}
                                    onClick={cycleSpeed}
                                >
                                    {speed}
                                </button>
                                <button
                                    type='button'
                                    className='flex items-center gap-1 rounded border border-border bg-bg-surface px-[7px] py-0.5 font-mono text-[11px] text-text-2 cursor-pointer hover:text-text-1 transition-colors disabled:opacity-50 disabled:cursor-not-allowed'
                                    title={
                                        downloading
                                            ? 'Downloading…'
                                            : 'Download full debate audio (.wav)'
                                    }
                                    aria-label={
                                        downloading ? 'Downloading…' : 'Download full debate audio'
                                    }
                                    disabled={downloading}
                                    onClick={() => {
                                        void handleDownload();
                                    }}
                                >
                                    {downloading ? (
                                        <svg
                                            viewBox='0 0 24 24'
                                            fill='none'
                                            stroke='currentColor'
                                            strokeWidth='2'
                                            strokeLinecap='round'
                                            className='w-[13px] h-[13px] animate-spin'
                                            aria-hidden
                                        >
                                            <circle cx='12' cy='12' r='9' strokeOpacity='0.25' />
                                            <path d='M12 3a9 9 0 0 1 9 9' />
                                        </svg>
                                    ) : (
                                        <IconDownload />
                                    )}
                                </button>
                            </div>
                        </div>

                        <Chapters debate={debate} />
                    </div>

                    {/* Right column: Audio info + Agent profiles */}
                    <div>
                        <AudioInfoCard debate={debate} />
                        <AgentProfilesCard debate={debate} />
                    </div>
                </div>
            </div>
        </div>
    );
}

export { AudioErrorBoundary as ErrorBoundary } from './ErrorBoundary';
