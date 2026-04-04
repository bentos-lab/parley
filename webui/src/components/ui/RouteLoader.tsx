export function RouteLoader() {
    return (
        <div className='fixed inset-0 z-50 flex flex-col items-center justify-center bg-(--bg-base)'>
            <span
                className='animate-pulse text-4xl tracking-wide text-(--text-1)'
                style={{ fontFamily: 'var(--font-display)' }}
            >
                parley
            </span>
            <div className='mt-6 flex gap-1.5'>
                <span className='h-1.5 w-1.5 animate-[bounce_1s_ease-in-out_0ms_infinite] rounded-full bg-accent' />
                <span className='h-1.5 w-1.5 animate-[bounce_1s_ease-in-out_150ms_infinite] rounded-full bg-accent' />
                <span className='h-1.5 w-1.5 animate-[bounce_1s_ease-in-out_300ms_infinite] rounded-full bg-accent' />
            </div>
        </div>
    );
}
