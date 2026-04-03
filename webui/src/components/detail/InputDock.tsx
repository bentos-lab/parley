import { Button } from '@/components/ui/Button';

export interface InputDockProps {
    currentRound: number;
    roundLimit: number;
    roundDuration: number;
    roundTimeRemaining: number;
    turnsInRound: number;
    sessionActive?: boolean;
    generating?: boolean;
    countdownRemaining?: number | null;
    onStart?: () => void;
    onStartNow?: () => void;
    onStop?: () => void;
    nextAgentName?: string;
    sessionComplete?: boolean;
}

function formatTime(seconds: number) {
    const m = Math.floor(seconds / 60);
    const s = seconds % 60;
    return m > 0 ? `${m}:${s.toString().padStart(2, '0')}` : `${s}s`;
}

export function InputDock({
    currentRound,
    roundLimit,
    roundDuration,
    roundTimeRemaining,
    turnsInRound,
    sessionActive = false,
    generating = false,
    countdownRemaining = null,
    onStart,
    onStartNow,
    onStop,
    nextAgentName,
    sessionComplete = false,
}: InputDockProps) {
    const roundsRemaining = Math.max(roundLimit - currentRound + 1, 0);
    const timerProgress =
        roundDuration > 0
            ? Math.min(
                  100,
                  Math.round(((roundDuration - roundTimeRemaining) / roundDuration) * 100),
              )
            : 0;
    const overallProgress =
        roundLimit > 0 ? Math.min(100, Math.round(((currentRound - 1) / roundLimit) * 100)) : 0;

    let title = 'Conversation paused';
    let description = 'Resume the session to start the next timed round.';

    if (sessionComplete) {
        title = 'All rounds complete';
        description = `The session finished ${roundLimit} timed rounds.`;
    } else if (generating) {
        title = `Round ${currentRound} — ${nextAgentName ?? 'Agent'} is speaking`;
        description = `${formatTime(roundTimeRemaining)} remaining · ${turnsInRound} turn${turnsInRound !== 1 ? 's' : ''} so far this round`;
    } else if (countdownRemaining !== null) {
        title = `Next round starts in ${countdownRemaining}s`;
        description = `Round ${currentRound} of ${roundLimit} is queued. Start immediately or wait for the countdown.`;
    } else if (sessionActive) {
        title = `Starting round ${currentRound}`;
        description = `${nextAgentName ?? 'Next agent'} will open. Time limit: ${formatTime(roundDuration)} per round.`;
    }

    return (
        <div className='shrink-0 border-t border-border bg-bg-panel px-4 py-3.5'>
            <div className='flex flex-col gap-3 rounded-xl border border-border bg-bg-surface px-4 py-3'>
                <div className='flex flex-col gap-3 md:flex-row md:items-end md:justify-between'>
                    <div>
                        <p className='font-mono text-[10px] uppercase tracking-[0.18em] text-text-3'>
                            Live session
                        </p>
                        <h2 className='mt-1 font-display text-lg text-text-1'>{title}</h2>
                        <p className='mt-2 max-w-2xl text-xs leading-6 text-text-2'>
                            {description}
                        </p>
                    </div>

                    <div className='flex flex-wrap items-center gap-2'>
                        <div className='rounded-full border border-border bg-bg-panel px-3 py-1 font-mono text-[10px] uppercase tracking-[0.12em] text-text-2'>
                            Round {Math.min(currentRound, roundLimit)}/{roundLimit}
                        </div>

                        {generating && roundTimeRemaining > 0 ? (
                            <div className='rounded-full border border-accent/30 bg-accent/5 px-3 py-1 font-mono text-[10px] uppercase tracking-[0.12em] text-accent'>
                                {formatTime(roundTimeRemaining)}
                            </div>
                        ) : null}

                        {generating ? (
                            <Button
                                type='button'
                                variant='secondary'
                                onClick={onStop}
                                className='h-8 px-3 text-[11px]'
                            >
                                Stop stream
                            </Button>
                        ) : null}

                        {!generating && countdownRemaining !== null && !sessionComplete ? (
                            <Button
                                type='button'
                                variant='accent'
                                onClick={onStartNow}
                                className='h-8 px-3 text-[11px]'
                            >
                                Start now
                            </Button>
                        ) : null}

                        {!generating &&
                        countdownRemaining === null &&
                        !sessionActive &&
                        !sessionComplete ? (
                            <Button
                                type='button'
                                variant='accent'
                                onClick={onStart}
                                className='h-8 px-3 text-[11px]'
                                data-testid='session-start-button'
                            >
                                {currentRound > 1 ? 'Resume session' : 'Start session'}
                            </Button>
                        ) : null}
                    </div>
                </div>

                {/* Round timer bar (within a round) */}
                {generating ||
                (sessionActive && countdownRemaining === null && !sessionComplete) ? (
                    <div>
                        <div className='mb-2 flex items-center justify-between text-[11px] text-text-3'>
                            <span>
                                Round {currentRound} · {turnsInRound} turn
                                {turnsInRound !== 1 ? 's' : ''}
                            </span>
                            <span className='font-mono'>
                                {formatTime(roundTimeRemaining)} / {formatTime(roundDuration)}
                            </span>
                        </div>
                        <div className='h-1.5 overflow-hidden rounded-full bg-bg-panel'>
                            <div
                                className='h-full rounded-full bg-accent transition-[width] duration-1000 ease-linear'
                                style={{ width: `${timerProgress}%` }}
                            />
                        </div>
                    </div>
                ) : null}

                {/* Overall session progress */}
                <div>
                    <div className='mb-2 flex items-center justify-between text-[11px] text-text-3'>
                        <span>
                            {sessionComplete
                                ? 'All rounds complete'
                                : `${roundsRemaining} round${roundsRemaining !== 1 ? 's' : ''} remaining`}
                        </span>
                        <span className='font-mono'>{formatTime(roundDuration)} per round</span>
                    </div>
                    <div className='h-1.5 overflow-hidden rounded-full bg-bg-panel'>
                        <div
                            className='h-full rounded-full bg-text-3/30 transition-[width] duration-300 ease-out'
                            style={{ width: `${overallProgress}%` }}
                        />
                    </div>
                </div>
            </div>
        </div>
    );
}
