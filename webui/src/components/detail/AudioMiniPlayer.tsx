/**
 * Mini audio player for the detail page status bar.
 * Shows the currently playing agent, round number, elapsed/duration, queue count,
 * and transport controls (previous, play/pause, next).
 */

function formatTime(seconds: number | null): string {
    if (seconds == null || isNaN(seconds)) return '0:00';
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
}

export interface AudioMiniPlayerProps {
    agentName: string;
    roundNumber: number;
    currentTime: number;
    duration: number | null;
    isPlaying: boolean;
    isLoading: boolean;
    queueCount: number;
    onTogglePlay: () => void;
    onNext: () => void;
    onPrevious: () => void;
    hasPrevious: boolean;
    hasNext: boolean;
}

export function AudioMiniPlayer({
    agentName,
    roundNumber,
    currentTime,
    duration,
    isPlaying,
    isLoading,
    queueCount,
    onTogglePlay,
    onNext,
    onPrevious,
    hasPrevious,
    hasNext,
}: AudioMiniPlayerProps) {
    return (
        <div className='flex w-full items-center gap-3'>
            {/* Agent & round info */}
            <div className='flex items-center gap-2 min-w-0 flex-1'>
                <span className='h-2 w-2 rounded-full bg-accent shrink-0 animate-pulse' />
                <div className='min-w-0 flex-1'>
                    <div className='flex items-center gap-2 min-w-0'>
                        <span className='text-xs font-medium text-text-1 truncate'>
                            {agentName}
                        </span>
                        <span className='text-xs text-text-3'>·</span>
                        <span className='text-xs text-text-2 whitespace-nowrap'>
                            Round {roundNumber}
                        </span>
                    </div>
                    {isLoading ? (
                        <div className='text-[11px] text-text-3 truncate'>
                            Preparing {agentName} speech...
                        </div>
                    ) : null}
                </div>
            </div>

            {/* Time display */}
            <div className='text-xs text-text-3 font-mono whitespace-nowrap'>
                {formatTime(currentTime)} / {formatTime(duration)}
            </div>

            {/* Queue indicator */}
            {queueCount > 0 && (
                <span className='text-xs text-text-3 bg-bg-elevated px-1.5 py-0.5 rounded'>
                    +{queueCount}
                </span>
            )}

            {/* Transport controls */}
            <div className='flex items-center gap-1'>
                <button
                    type='button'
                    onClick={onPrevious}
                    disabled={!hasPrevious}
                    className='p-1 rounded hover:bg-bg-hover disabled:opacity-30 disabled:cursor-not-allowed transition-colors'
                    aria-label='Previous round'
                >
                    <svg
                        className='w-3.5 h-3.5 text-text-2'
                        fill='currentColor'
                        viewBox='0 0 24 24'
                    >
                        <path d='M6 6h2v12H6V6zm3.5 6l8.5 6V6l-8.5 6z' />
                    </svg>
                </button>

                <button
                    type='button'
                    onClick={onTogglePlay}
                    disabled={isLoading}
                    className='p-1.5 rounded-full bg-accent text-white hover:opacity-90 disabled:opacity-50 transition-opacity'
                    aria-label={isPlaying ? 'Pause' : 'Play'}
                >
                    {isLoading ? (
                        <svg className='w-3.5 h-3.5 animate-spin' fill='none' viewBox='0 0 24 24'>
                            <circle
                                className='opacity-25'
                                cx='12'
                                cy='12'
                                r='10'
                                stroke='currentColor'
                                strokeWidth='4'
                            />
                            <path
                                className='opacity-75'
                                fill='currentColor'
                                d='M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z'
                            />
                        </svg>
                    ) : isPlaying ? (
                        <svg className='w-3.5 h-3.5' fill='currentColor' viewBox='0 0 24 24'>
                            <path d='M6 4h4v16H6V4zm8 0h4v16h-4V4z' />
                        </svg>
                    ) : (
                        <svg className='w-3.5 h-3.5' fill='currentColor' viewBox='0 0 24 24'>
                            <path d='M8 5v14l11-7z' />
                        </svg>
                    )}
                </button>

                <button
                    type='button'
                    onClick={onNext}
                    disabled={!hasNext}
                    className='p-1 rounded hover:bg-bg-hover disabled:opacity-30 disabled:cursor-not-allowed transition-colors'
                    aria-label='Next round'
                >
                    <svg
                        className='w-3.5 h-3.5 text-text-2'
                        fill='currentColor'
                        viewBox='0 0 24 24'
                    >
                        <path d='M6 18l8.5-6L6 6v12zm8.5 0V6h2v12h-2z' />
                    </svg>
                </button>
            </div>
        </div>
    );
}
