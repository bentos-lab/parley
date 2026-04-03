import { useStoreState } from '@/app/store/hooks';

export interface SSEStatusBarProps {
    activeRoute: string;
    statusText?: string;
}

export function SSEStatusBar({ activeRoute, statusText }: SSEStatusBarProps) {
    const sseActive = useStoreState((store) => store.sessionUI.sseActive);
    const isLive = sseActive && Boolean(activeRoute);

    if (!isLive && !statusText) return null;

    return (
        <div
            className='flex shrink-0 items-center gap-1.5 border-b border-border bg-bg-panel px-5 py-1.5 text-[10px] text-text-3'
            data-testid='sse-status-bar'
        >
            <span
                className={`h-[5px] w-[5px] rounded-full ${isLive ? 'animate-[pulse_1.5s_infinite]' : ''}`}
                style={{ backgroundColor: isLive ? '#4e8f68' : 'var(--text-3)' }}
                aria-hidden='true'
            />
            <span className='font-mono'>
                {statusText ?? (isLive ? activeRoute : 'Idle — waiting for generation')}
            </span>
        </div>
    );
}
