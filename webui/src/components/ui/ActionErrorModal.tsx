import { Button } from './Button';

export interface ActionErrorModalProps {
    open: boolean;
    title: string;
    message: string;
    supportingText?: string;
    onClose: () => void;
    primaryActionLabel?: string;
    onPrimaryAction?: () => void;
}

export function ActionErrorModal({
    open,
    title,
    message,
    supportingText,
    onClose,
    primaryActionLabel,
    onPrimaryAction,
}: ActionErrorModalProps) {
    if (!open) {
        return null;
    }

    return (
        <div
            className='fixed inset-0 z-50 flex items-center justify-center bg-black/55 px-4 backdrop-blur-[2px]'
            onClick={onClose}
        >
            <div
                className='w-full max-w-md rounded-2xl border border-border bg-bg-panel p-6 shadow-2xl'
                onClick={(event) => event.stopPropagation()}
                role='alertdialog'
                aria-modal='true'
                aria-labelledby='action-error-title'
                aria-describedby='action-error-description'
            >
                <div className='flex h-10 w-10 items-center justify-center rounded-full border border-error/25 bg-error/10 text-error'>
                    <svg
                        viewBox='0 0 16 16'
                        className='h-4.5 w-4.5 fill-none'
                        stroke='currentColor'
                    >
                        <path
                            d='M8 4.2v4.1M8 11.3h.01M8.86 1.93l5.9 10.2A1 1 0 0 1 13.9 13.6H2.1a1 1 0 0 1-.86-1.47l5.9-10.2a1 1 0 0 1 1.72 0Z'
                            strokeWidth='1.25'
                            strokeLinecap='round'
                            strokeLinejoin='round'
                        />
                    </svg>
                </div>

                <p className='mt-4 font-mono text-[10px] uppercase tracking-[0.16em] text-text-3'>
                    Request failed
                </p>
                <h2 id='action-error-title' className='mt-2 font-display text-2xl text-text-1'>
                    {title}
                </h2>
                <p id='action-error-description' className='mt-3 text-sm leading-6 text-text-2'>
                    {message}
                </p>

                {supportingText ? (
                    <p className='mt-3 rounded-xl border border-border bg-bg-surface px-4 py-3 text-xs leading-5 text-text-3'>
                        {supportingText}
                    </p>
                ) : null}

                <div className='mt-6 flex flex-wrap justify-end gap-3'>
                    <Button type='button' variant='ghost' onClick={onClose}>
                        Dismiss
                    </Button>
                    {primaryActionLabel && onPrimaryAction ? (
                        <Button type='button' variant='danger' onClick={onPrimaryAction}>
                            {primaryActionLabel}
                        </Button>
                    ) : null}
                </div>
            </div>
        </div>
    );
}
