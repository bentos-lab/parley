import { useEffect, useRef, useState } from 'react';
import type { LoaderFunctionArgs } from 'react-router-dom';
import { useLoaderData, useNavigate, useSearchParams } from 'react-router-dom';
import { AgentPanel } from '@/components/detail/AgentPanel';
import { AudioMiniPlayer } from '@/components/detail/AudioMiniPlayer';
import { RoundList } from '@/components/detail/RoundList';
import { ActionErrorModal } from '@/components/ui/ActionErrorModal';
import { useAutoScroll } from '@/hooks/useAutoScroll';
import {
    clearDebateAudioCache,
    useAudioQueueState,
    useRoundAudioCache,
    useRoundAudioStatus,
} from '@/hooks/useRoundAudioCache';
import { useRoundGeneration } from '@/hooks/useRoundGeneration';
import {
    DEFAULT_TURN_COUNT,
    readStoredSessionConfig,
    writeStoredSessionConfig,
} from '@/lib/debateSessionConfig';
import { getDebate } from '@/services/api/debates';
import { getErrorMessage } from '@/services/api/http';
import { DebateSchema } from '@/services/api/schemas';
import type { Debate, Round } from '@/types';

// eslint-disable-next-line react-refresh/only-export-components
export async function loader({ params }: LoaderFunctionArgs) {
    const data = await getDebate(params.debateId!);
    // Real BE may not echo `id` in the response body — inject it from the URL param
    return DebateSchema.parse({ id: params.debateId, ...data });
}

export function Component() {
    const debate = useLoaderData() as Debate;
    const navigate = useNavigate();
    const [searchParams, setSearchParams] = useSearchParams();
    const [isMobileAgentPanelOpen, setIsMobileAgentPanelOpen] = useState(false);
    const roundParam = searchParams.get('round');
    const autoplayParam = searchParams.get('autoplay');

    // Turn count — Priority: URL param > localStorage > default
    const [activeTurnCount, setActiveTurnCount] = useState(() => {
        if (roundParam) {
            const parsed = Number.parseInt(roundParam, 10);
            if (!Number.isNaN(parsed) && parsed >= 1 && parsed <= 20) {
                return parsed;
            }
        }
        return readStoredSessionConfig(debate.id)?.turnCount ?? DEFAULT_TURN_COUNT;
    });

    // Immediate audio playback toggle — Priority: URL param > false
    const [playAudioImmediately, setPlayAudioImmediately] = useState(() => autoplayParam === '1');

    // Track whether generation session has completed
    const [sessionCompleted, setSessionCompleted] = useState(false);

    // Track which agent is highlighted (null = none)
    const [highlightedAgentId, setHighlightedAgentId] = useState<string | null>(null);

    // Accumulate rounds generated in current session
    const [generatedRounds, setGeneratedRounds] = useState<Round[]>([]);
    const [generationError, setGenerationError] = useState<string | null>(null);

    // Audio playback via shared cache — use queue for sequential autoplay
    const {
        enqueue,
        stop: stopAudio,
        togglePause,
        next: nextAudio,
        previous: prevAudio,
    } = useRoundAudioCache();
    const queueState = useAudioQueueState();
    const playAudioImmediatelyRef = useRef(playAudioImmediately);
    useEffect(() => {
        playAudioImmediatelyRef.current = playAudioImmediately;
    }, [playAudioImmediately]);

    useEffect(() => {
        return () => {
            stopAudio();
            clearDebateAudioCache(debate.id);
        };
    }, [debate.id, stopAudio]);

    // Track index of last round that triggered auto-enqueue (to avoid duplicates on re-render)
    const lastEnqueuedIndexRef = useRef<number | null>(null);

    // Clear generated rounds when the debate revalidates with new persisted rounds
    const prevRoundsLengthRef = useRef(debate.rounds.length);
    useEffect(() => {
        if (debate.rounds.length !== prevRoundsLengthRef.current) {
            prevRoundsLengthRef.current = debate.rounds.length;
            setGeneratedRounds([]);
        }
    }, [debate.rounds.length]);

    const scrollRef = useRef<HTMLDivElement>(null);
    const hasAutoStarted = useRef(false);

    const {
        isGenerating,
        isActive,
        isStopping,
        generatedCount,
        lastGeneratedRoundIndex,
        start: startGeneration,
        stop,
    } = useRoundGeneration({
        debateId: debate.id,
        onRound: (round) => {
            setGeneratedRounds((prev) => [...prev, round]);
        },
        onError: (err) => {
            setGenerationError(getErrorMessage(err, 'Failed to generate the next round.'));
        },
    });

    // Auto-enqueue audio when a new round completes (if enabled)
    useEffect(() => {
        if (
            playAudioImmediatelyRef.current &&
            lastGeneratedRoundIndex != null &&
            lastGeneratedRoundIndex !== lastEnqueuedIndexRef.current
        ) {
            lastEnqueuedIndexRef.current = lastGeneratedRoundIndex;
            enqueue(debate.id, lastGeneratedRoundIndex);
        }
    }, [lastGeneratedRoundIndex, debate.id, enqueue]);

    // Get audio status for the currently playing round (for mini-player metadata)
    const currentRoundIndex = queueState.current?.roundIndex ?? null;
    const currentAudioStatus = useRoundAudioStatus(
        queueState.current?.debateId,
        currentRoundIndex ?? undefined,
    );

    // Find agent info for the currently playing round
    const allRounds = [...debate.rounds, ...generatedRounds];
    const currentPlayingRound = currentRoundIndex != null ? allRounds[currentRoundIndex] : null;
    const currentPlayingAgent = currentPlayingRound
        ? debate.agents.find((a) => a.id === currentPlayingRound.agent_id)
        : null;

    // Start generation with current settings
    const handleStart = () => {
        const startingIndex = debate.rounds.length + generatedRounds.length;
        writeStoredSessionConfig(debate.id, { turnCount: activeTurnCount });
        setSessionCompleted(false);
        setGenerationError(null);
        lastEnqueuedIndexRef.current = null;
        startGeneration(activeTurnCount, startingIndex);
    };

    // Stop generation AND audio playback
    const handleStop = () => {
        stop();
        stopAudio();
    };

    // Mark session complete when generation finishes
    useEffect(() => {
        if (!isActive && generatedCount > 0) {
            setSessionCompleted(true);
        }
    }, [isActive, generatedCount]);

    useEffect(() => {
        const mediaQuery = window.matchMedia('(min-width: 1024px)');
        const handleChange = (event: MediaQueryListEvent | MediaQueryList) => {
            if (event.matches) {
                setIsMobileAgentPanelOpen(false);
            }
        };

        handleChange(mediaQuery);

        if (typeof mediaQuery.addEventListener === 'function') {
            mediaQuery.addEventListener('change', handleChange);
            return () => mediaQuery.removeEventListener('change', handleChange);
        }

        mediaQuery.addListener(handleChange);
        return () => mediaQuery.removeListener(handleChange);
    }, []);

    const autoScroll = useAutoScroll({
        containerRef: scrollRef,
        trigger: generatedRounds.length + (isGenerating ? 1 : 0),
    });

    // Auto-start generation when redirected from create with ?round= param
    useEffect(() => {
        if (hasAutoStarted.current) return;
        if (debate.rounds.length === 0 && roundParam) {
            hasAutoStarted.current = true;
            // Clear the search param so refresh doesn't re-trigger
            setSearchParams({}, { replace: true });
            handleStart();
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [debate.rounds.length, roundParam, setSearchParams]);

    function handleStartNext() {
        setGeneratedRounds([]);
        setSessionCompleted(false);
        lastEnqueuedIndexRef.current = null;
        stopAudio();
        handleStart();
    }

    // Determine what to show in status bar
    const hasActivePlayer = queueState.playerStatus !== 'idle' && currentPlayingAgent;
    const showMiniPlayer = Boolean(hasActivePlayer);
    const showGeneratingStatus = isActive && !showMiniPlayer;

    return (
        <section className='flex h-full overflow-hidden bg-bg-base'>
            {/* Rounds column */}
            <div className='flex flex-1 flex-col overflow-hidden border-r border-border'>
                {/* Status bar with mini-player */}
                {(showGeneratingStatus || showMiniPlayer) && (
                    <div className='flex items-center gap-2 border-b border-border bg-bg-panel px-4 py-2'>
                        {showMiniPlayer && currentPlayingAgent ? (
                            <AudioMiniPlayer
                                agentName={currentPlayingAgent.name}
                                roundNumber={(currentRoundIndex ?? 0) + 1}
                                currentTime={currentAudioStatus.currentTime}
                                duration={currentAudioStatus.duration}
                                isPlaying={queueState.playerStatus === 'playing'}
                                isLoading={queueState.playerStatus === 'loading'}
                                queueCount={queueState.queue.length}
                                onTogglePlay={togglePause}
                                onNext={nextAudio}
                                onPrevious={prevAudio}
                                hasPrevious={
                                    queueState.history.length > 0 || (currentRoundIndex ?? 0) > 0
                                }
                                hasNext={queueState.queue.length > 0}
                            />
                        ) : (
                            <>
                                {isActive && (
                                    <span className='h-2 w-2 rounded-full bg-accent animate-pulse' />
                                )}
                                <span className='text-xs text-text-2'>
                                    {isGenerating
                                        ? 'Generating round…'
                                        : `Generated ${generatedCount} of ${activeTurnCount} rounds`}
                                </span>
                            </>
                        )}
                    </div>
                )}

                <RoundList
                    debate={debate}
                    scrollRef={scrollRef}
                    streamingRounds={generatedRounds}
                    isThinking={isGenerating}
                    autoScrollPaused={autoScroll.paused}
                    onScrollToBottom={autoScroll.resumeNow}
                    highlightedAgentId={highlightedAgentId}
                />
            </div>

            {/* Right side panel */}
            <AgentPanel
                debate={debate}
                streamingCount={generatedRounds.length}
                isStreaming={isActive}
                isStopping={isStopping}
                sseCompleted={sessionCompleted}
                turnCount={activeTurnCount}
                onTurnCountChange={setActiveTurnCount}
                playAudioImmediately={playAudioImmediately}
                onPlayAudioChange={setPlayAudioImmediately}
                onStart={handleStart}
                onStop={handleStop}
                onStartNext={handleStartNext}
                highlightedAgentId={highlightedAgentId}
                onAgentHighlight={setHighlightedAgentId}
                mobileOpen={isMobileAgentPanelOpen}
                onMobileToggle={() => setIsMobileAgentPanelOpen((prev) => !prev)}
                onMobileClose={() => setIsMobileAgentPanelOpen(false)}
            />

            <ActionErrorModal
                open={generationError !== null}
                title='Round generation failed'
                message={generationError ?? ''}
                supportingText='If you recently changed your LLM or TTS credentials, review them in Settings and try generating again.'
                onClose={() => setGenerationError(null)}
                primaryActionLabel='Open settings'
                onPrimaryAction={() => navigate('/settings')}
            />
        </section>
    );
}

export { DetailErrorBoundary as ErrorBoundary } from './ErrorBoundary';
