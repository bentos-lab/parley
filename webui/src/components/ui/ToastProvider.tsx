import {
    createContext,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useState,
    type ReactNode,
} from 'react';

type ToastVariant = 'error' | 'success' | 'info';

interface ToastRecord {
    id: string;
    title?: string;
    message: string;
    variant: ToastVariant;
    duration: number;
}

interface ShowToastInput {
    title?: string;
    message: string;
    duration?: number;
    id?: string;
}

interface ToastContextValue {
    show: (toast: ShowToastInput & { variant?: ToastVariant }) => string;
    error: (message: string, options?: Omit<ShowToastInput, 'message'>) => string;
    success: (message: string, options?: Omit<ShowToastInput, 'message'>) => string;
    info: (message: string, options?: Omit<ShowToastInput, 'message'>) => string;
    dismiss: (id: string) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

function createToastId() {
    return globalThis.crypto?.randomUUID?.() ?? `toast-${Date.now()}-${Math.random()}`;
}

function variantClasses(variant: ToastVariant) {
    switch (variant) {
        case 'error':
            return {
                panel: 'border-error/40 bg-error-bg/95 text-text-1',
                icon: 'bg-error/15 text-error border-error/30',
                label: 'Problem',
            };
        case 'success':
            return {
                panel: 'border-agent-1-border bg-bg-panel/95 text-text-1',
                icon: 'bg-agent-1-bg text-agent-1 border-agent-1-border',
                label: 'Done',
            };
        default:
            return {
                panel: 'border-border-hi bg-bg-panel/95 text-text-1',
                icon: 'bg-accent-bg text-accent border-accent/30',
                label: 'Notice',
            };
    }
}

function ToastItem({ toast, onDismiss }: { toast: ToastRecord; onDismiss: (id: string) => void }) {
    useEffect(() => {
        const timer = window.setTimeout(() => onDismiss(toast.id), toast.duration);
        return () => window.clearTimeout(timer);
    }, [onDismiss, toast.duration, toast.id]);

    const styles = variantClasses(toast.variant);

    return (
        <div
            className={`pointer-events-auto w-full max-w-sm rounded-2xl border p-3.5 shadow-2xl backdrop-blur-sm ${styles.panel} animate-[toastEnter_180ms_ease-out]`}
            role={toast.variant === 'error' ? 'alert' : 'status'}
            aria-live={toast.variant === 'error' ? 'assertive' : 'polite'}
        >
            <div className='flex items-start gap-3'>
                <div
                    className={`mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full border ${styles.icon}`}
                    aria-hidden='true'
                >
                    {toast.variant === 'error' ? (
                        <svg
                            viewBox='0 0 16 16'
                            className='h-4 w-4 fill-none'
                            stroke='currentColor'
                        >
                            <path
                                d='M8 4.2v4.1M8 11.3h.01M8.86 1.93l5.9 10.2A1 1 0 0 1 13.9 13.6H2.1a1 1 0 0 1-.86-1.47l5.9-10.2a1 1 0 0 1 1.72 0Z'
                                strokeWidth='1.25'
                                strokeLinecap='round'
                                strokeLinejoin='round'
                            />
                        </svg>
                    ) : toast.variant === 'success' ? (
                        <svg
                            viewBox='0 0 16 16'
                            className='h-4 w-4 fill-none'
                            stroke='currentColor'
                        >
                            <path
                                d='M3.2 8.4 6.5 11.5 12.8 4.9'
                                strokeWidth='1.4'
                                strokeLinecap='round'
                                strokeLinejoin='round'
                            />
                        </svg>
                    ) : (
                        <svg
                            viewBox='0 0 16 16'
                            className='h-4 w-4 fill-none'
                            stroke='currentColor'
                        >
                            <path
                                d='M8 4.3h.01M7.4 6.6h1.2v5H7.4zM8 14A6 6 0 1 0 8 2a6 6 0 0 0 0 12Z'
                                strokeWidth='1.25'
                                strokeLinecap='round'
                                strokeLinejoin='round'
                            />
                        </svg>
                    )}
                </div>

                <div className='min-w-0 flex-1'>
                    <p className='font-mono text-[10px] uppercase tracking-[0.16em] text-text-3'>
                        {toast.title ?? styles.label}
                    </p>
                    <p className='mt-1 text-sm leading-5 text-text-1'>{toast.message}</p>
                </div>

                <button
                    type='button'
                    onClick={() => onDismiss(toast.id)}
                    className='mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-text-3 transition-colors hover:bg-bg-hover hover:text-text-1 cursor-pointer'
                    aria-label='Dismiss notification'
                >
                    <svg
                        viewBox='0 0 16 16'
                        className='h-3.5 w-3.5 fill-none'
                        stroke='currentColor'
                    >
                        <path d='M4 4l8 8M12 4 4 12' strokeWidth='1.35' strokeLinecap='round' />
                    </svg>
                </button>
            </div>
        </div>
    );
}

export function ToastProvider({ children }: { children: ReactNode }) {
    const [toasts, setToasts] = useState<ToastRecord[]>([]);

    const dismiss = useCallback((id: string) => {
        setToasts((current) => current.filter((toast) => toast.id !== id));
    }, []);

    const show = useCallback((toast: ShowToastInput & { variant?: ToastVariant }) => {
        const id = toast.id ?? createToastId();
        const variant = toast.variant ?? 'info';
        const duration = toast.duration ?? (variant === 'error' ? 6000 : 4200);

        setToasts((current) => {
            const nextToast: ToastRecord = {
                id,
                title: toast.title,
                message: toast.message,
                variant,
                duration,
            };

            return [...current.filter((item) => item.id !== id), nextToast];
        });

        return id;
    }, []);

    const value = useMemo<ToastContextValue>(
        () => ({
            show,
            dismiss,
            error: (message, options) => show({ ...options, message, variant: 'error' }),
            success: (message, options) => show({ ...options, message, variant: 'success' }),
            info: (message, options) => show({ ...options, message, variant: 'info' }),
        }),
        [dismiss, show],
    );

    return (
        <ToastContext.Provider value={value}>
            {children}
            <div className='pointer-events-none fixed right-4 top-4 z-[60] flex w-[min(100%-2rem,24rem)] flex-col gap-3 sm:right-5 sm:top-5'>
                {toasts.map((toast) => (
                    <ToastItem key={toast.id} toast={toast} onDismiss={dismiss} />
                ))}
            </div>
        </ToastContext.Provider>
    );
}

// eslint-disable-next-line react-refresh/only-export-components
export function useToast() {
    const context = useContext(ToastContext);

    if (!context) {
        throw new Error('useToast must be used within ToastProvider');
    }

    return context;
}
