import type { DebateSummary } from '@/types';

export interface DeleteConfirmModalProps {
    debate: DebateSummary;
    onConfirm: () => void;
    onCancel: () => void;
    deleting: boolean;
}

export function DeleteConfirmModal({
    debate,
    onConfirm,
    onCancel,
    deleting,
}: DeleteConfirmModalProps) {
    const titleId = `delete-debate-title-${debate.id}`;
    const descriptionId = `delete-debate-description-${debate.id}`;

    return (
        <div
            className='fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-[2px]'
            onClick={onCancel}
        >
            <div
                className='mx-4 w-full max-w-sm rounded-xl border border-border bg-bg-panel p-5 shadow-xl'
                onClick={(e) => e.stopPropagation()}
                role='dialog'
                aria-modal='true'
                aria-labelledby={titleId}
                aria-describedby={descriptionId}
            >
                <div className='mb-3 flex h-9 w-9 items-center justify-center rounded-full bg-red-500/10'>
                    <svg width='16' height='16' viewBox='0 0 16 16' fill='none' aria-hidden='true'>
                        <path
                            d='M2 4h12M5 4V2.5A.5.5 0 0 1 5.5 2h5a.5.5 0 0 1 .5.5V4M6 7v4M10 7v4M3 4l1 9.5A.5.5 0 0 0 4.5 14h7a.5.5 0 0 0 .5-.5L13 4'
                            stroke='#ef4444'
                            strokeWidth='1.2'
                            strokeLinecap='round'
                            strokeLinejoin='round'
                        />
                    </svg>
                </div>

                <h2 id={titleId} className='mb-1 text-sm font-semibold text-text-1'>
                    Delete debate?
                </h2>
                <p id={descriptionId} className='mb-4 text-[12px] leading-[1.6] text-text-3'>
                    <span className='font-medium text-text-2'>{debate.name}</span> will be
                    permanently deleted from disk. This cannot be undone.
                </p>

                <div className='flex items-center justify-end gap-2'>
                    <button
                        type='button'
                        onClick={onCancel}
                        disabled={deleting}
                        className='cursor-pointer rounded-md border border-border px-3.5 py-1.5 text-xs font-medium text-text-2 transition-colors hover:bg-bg-hover disabled:opacity-40'
                    >
                        Cancel
                    </button>
                    <button
                        type='button'
                        onClick={onConfirm}
                        disabled={deleting}
                        className='flex cursor-pointer items-center gap-1.5 rounded-md bg-red-600 px-3.5 py-1.5 text-xs font-medium text-white transition-opacity hover:opacity-90 disabled:opacity-60'
                    >
                        {deleting && (
                            <span className='h-3 w-3 animate-spin rounded-full border border-white/40 border-t-white' />
                        )}
                        Delete
                    </button>
                </div>
            </div>
        </div>
    );
}
