import { WhatsAppConnectCard } from '@/components/settings/WhatsAppConnectCard';

export function Component() {
    return (
        <div className='mx-auto flex w-full max-w-4xl flex-col gap-6 p-6 font-sans text-text-1 sm:p-8'>
            <div className='max-w-2xl'>
                <h1 className='mb-2 font-display text-2xl text-text-1'>WhatsApp Integration</h1>
                <p className='text-sm leading-6 text-text-3'>
                    Pair your WhatsApp account with Parley, scan a QR code from your phone, and
                    manage reconnects without leaving the app.
                </p>
            </div>

            <div className='grid gap-6 lg:grid-cols-[minmax(0,1fr)_300px]'>
                <WhatsAppConnectCard />

                <aside className='rounded-lg border border-border bg-bg-surface p-4'>
                    <h2 className='mb-3 text-sm font-semibold text-text-1'>How it works</h2>
                    <ol className='space-y-2 text-xs leading-5 text-text-3'>
                        <li>1. Start a new pairing session from this page.</li>
                        <li>2. Scan the QR code in WhatsApp on your phone.</li>
                        <li>3. Wait until your device finishes synchronizing.</li>
                        <li>4. Confirm completion here to close the connection flow.</li>
                    </ol>
                    <div className='mt-4 rounded-lg border border-border-mid bg-bg-base px-3 py-2'>
                        <p className='text-[11px] uppercase tracking-[0.08em] text-text-3'>Tips</p>
                        <p className='mt-1 text-xs leading-5 text-text-2'>
                            After pairing, open your own Saved Messages chat in WhatsApp and send a
                            `/parley ...` command there. Parley does not create a new conversation
                            thread for you.
                        </p>
                    </div>
                </aside>
            </div>
        </div>
    );
}

export function ErrorBoundary() {
    return (
        <div className='mx-auto w-full max-w-4xl p-6 font-sans sm:p-8'>
            <h1 className='mb-4 font-display text-2xl text-text-1'>WhatsApp Integration</h1>
            <div className='rounded-lg border border-error/30 bg-error/10 px-4 py-3 text-sm text-error'>
                Failed to load the WhatsApp integration page.
            </div>
        </div>
    );
}
