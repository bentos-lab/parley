import { useEffect, useState } from 'react';
import { QRCodeSVG } from 'qrcode.react';
import { useWhatsAppConnect } from '@/hooks/useWhatsAppConnect';
import { Button } from '@/components/ui/Button';

const scannedConfirmDelaySeconds = 5;

function WhatsAppIcon({ className }: { className?: string }) {
    return (
        <svg
            className={className ?? 'w-4 h-4'}
            viewBox='0 0 24 24'
            fill='currentColor'
            aria-hidden='true'
        >
            <path d='M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z' />
        </svg>
    );
}

function CheckCircleIcon({ className }: { className?: string }) {
    return (
        <svg
            className={className ?? 'w-4 h-4'}
            viewBox='0 0 24 24'
            fill='none'
            stroke='currentColor'
            strokeWidth='2'
            strokeLinecap='round'
            strokeLinejoin='round'
            aria-hidden='true'
        >
            <path d='M22 11.08V12a10 10 0 1 1-5.93-9.14' />
            <polyline points='22 4 12 14.01 9 11.01' />
        </svg>
    );
}

function RefreshIcon({ className }: { className?: string }) {
    return (
        <svg
            className={className ?? 'w-4 h-4'}
            viewBox='0 0 24 24'
            fill='none'
            stroke='currentColor'
            strokeWidth='2'
            strokeLinecap='round'
            strokeLinejoin='round'
            aria-hidden='true'
        >
            <polyline points='23 4 23 10 17 10' />
            <polyline points='1 20 1 14 7 14' />
            <path d='M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15' />
        </svg>
    );
}

function LoadingSpinner({ className }: { className?: string }) {
    return (
        <span
            className={`inline-block h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent ${className ?? ''}`}
            aria-hidden='true'
        />
    );
}

export function WhatsAppConnectCard() {
    const { state, countdown, connect, confirmSync, disconnect, retry, checkStatus } =
        useWhatsAppConnect();
    const [confirmCooldown, setConfirmCooldown] = useState({ active: false, remaining: 0 });

    useEffect(() => {
        void checkStatus();
    }, [checkStatus]);

    useEffect(() => {
        if (state.status !== 'scanned') {
            const resetId = window.setTimeout(() => {
                setConfirmCooldown({ active: false, remaining: 0 });
            }, 0);

            return () => {
                window.clearTimeout(resetId);
            };
        }

        const startedAt = Date.now();

        const updateCountdown = () => {
            const remaining = Math.max(
                0,
                scannedConfirmDelaySeconds - Math.floor((Date.now() - startedAt) / 1000),
            );

            setConfirmCooldown({ active: true, remaining });

            if (remaining <= 0) {
                window.clearInterval(intervalId);
            }
        };

        const kickoffId = window.setTimeout(updateCountdown, 0);
        const intervalId = window.setInterval(() => {
            updateCountdown();
        }, 1000);

        return () => {
            window.clearTimeout(kickoffId);
            window.clearInterval(intervalId);
        };
    }, [state.status]);

    let confirmCountdown = 0;

    if (state.status === 'scanned') {
        confirmCountdown = confirmCooldown.active
            ? confirmCooldown.remaining
            : scannedConfirmDelaySeconds;
    }

    const formatCountdown = (seconds: number) => {
        const mins = Math.floor(seconds / 60);
        const secs = seconds % 60;
        return mins > 0 ? `${mins}:${secs.toString().padStart(2, '0')}` : `${secs}s`;
    };

    return (
        <section className='rounded-lg border border-border bg-bg-surface p-4'>
            <h2 className='mb-4 flex items-center gap-2 text-sm font-semibold text-text-1'>
                <WhatsAppIcon className='h-4 w-4 opacity-70' />
                WhatsApp Integration
            </h2>

            {(state.status === 'idle' || state.status === 'checking') && (
                <div className='flex items-center gap-2 text-xs text-text-3'>
                    <LoadingSpinner className='h-3 w-3' />
                    Checking connection status...
                </div>
            )}

            {state.status === 'disconnected' && (
                <div className='space-y-3'>
                    <p className='text-xs text-text-2'>
                        Connect your WhatsApp to send <code className='text-accent'>/parley</code>{' '}
                        commands from your phone.
                    </p>
                    <Button variant='secondary' onClick={connect}>
                        <WhatsAppIcon className='h-3.5 w-3.5' />
                        Connect WhatsApp
                    </Button>
                </div>
            )}

            {state.status === 'connecting' && (
                <div className='flex items-center gap-2 text-xs text-text-3'>
                    <LoadingSpinner className='h-3 w-3' />
                    Initializing connection...
                </div>
            )}

            {state.status === 'qr' && (
                <div className='space-y-4'>
                    <div className='flex flex-col items-center gap-3'>
                        <div className='rounded-lg bg-white p-3'>
                            <QRCodeSVG
                                value={state.code}
                                size={160}
                                level='L'
                                includeMargin={false}
                            />
                        </div>
                        <div className='text-center'>
                            <p className='mb-1 text-xs text-text-2'>
                                Scan with WhatsApp to connect
                            </p>
                            <p
                                className={`text-xs font-mono ${countdown <= 10 ? 'text-error' : 'text-text-3'}`}
                            >
                                Expires in {formatCountdown(countdown)}
                            </p>
                        </div>
                    </div>
                    <ol className='list-inside list-decimal space-y-1 text-[11px] text-text-3'>
                        <li>Open WhatsApp on your phone</li>
                        <li>
                            Go to{' '}
                            <strong className='text-text-2'>Settings -&gt; Linked Devices</strong>
                        </li>
                        <li>
                            Tap <strong className='text-text-2'>Link a Device</strong>
                        </li>
                        <li>Point your phone at this QR code</li>
                    </ol>
                </div>
            )}

            {state.status === 'scanned' && (
                <div className='space-y-3'>
                    <div
                        className='flex items-center gap-2 text-xs text-accent'
                        role='status'
                        aria-live='polite'
                    >
                        <LoadingSpinner className='h-3 w-3' />
                        Synchronizing with WhatsApp...
                    </div>
                    <p
                        className={`rounded-md border px-3 py-2 text-[11px] ${confirmCountdown > 0 ? 'border-accent bg-accent/10 text-accent' : 'border-border-mid bg-bg-elevated text-text-2'}`}
                    >
                        {confirmCountdown > 0
                            ? `Wait ${confirmCountdown}s and only click Done after WhatsApp shows the device is linked on your phone.`
                            : 'Click Done only after WhatsApp shows the device is linked on your phone.'}
                    </p>
                    <Button variant='accent' onClick={confirmSync} disabled={confirmCountdown > 0}>
                        <CheckCircleIcon className='h-3.5 w-3.5' />
                        {confirmCountdown > 0 ? `Done (${confirmCountdown}s)` : 'Done'}
                    </Button>
                </div>
            )}

            {state.status === 'connected' && (
                <div className='space-y-3'>
                    <div
                        className='flex items-center gap-2 text-xs text-agent-1'
                        role='status'
                        aria-live='polite'
                    >
                        <CheckCircleIcon className='h-4 w-4' />
                        WhatsApp is connected
                    </div>
                    <p className='text-[11px] text-text-3'>
                        Send <code className='text-accent'>/parley</code> commands to your Saved
                        Messages chat.
                    </p>
                    <Button variant='ghost' onClick={() => void disconnect()}>
                        Disconnect
                    </Button>
                </div>
            )}

            {state.status === 'timeout' && (
                <div className='space-y-3'>
                    <div className='text-xs text-stale' role='alert'>
                        QR code expired
                    </div>
                    <Button variant='secondary' onClick={retry}>
                        <RefreshIcon className='h-3.5 w-3.5' />
                        Refresh Page
                    </Button>
                </div>
            )}

            {state.status === 'error' && (
                <div className='space-y-3'>
                    <div className='text-xs text-error' role='alert'>
                        {state.message}
                    </div>
                    <Button variant='secondary' onClick={retry}>
                        <RefreshIcon className='h-3.5 w-3.5' />
                        Refresh Page
                    </Button>
                </div>
            )}
        </section>
    );
}
