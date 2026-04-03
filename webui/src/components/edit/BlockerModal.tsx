import { Button } from '@/components/ui/Button';

export interface BlockerModalProps {
    open: boolean;
    onLeave: () => void;
    onStay: () => void;
}

export function BlockerModal({ open, onLeave, onStay }: BlockerModalProps) {
    if (!open) {
        return null;
    }

    return (
        <div className='fixed inset-0 z-40 flex items-center justify-center bg-bg-base/70 px-4 backdrop-blur-sm'>
            <div
                className='w-full max-w-md rounded-2xl border border-border bg-bg-panel p-6 shadow-2xl'
                role='dialog'
                aria-modal='true'
                aria-labelledby='edit-blocker-title'
            >
                <p className='font-mono text-[10px] uppercase tracking-[0.18em] text-text-3'>
                    Unsaved changes
                </p>
                <h2 id='edit-blocker-title' className='mt-2 font-display text-2xl text-text-1'>
                    Leave this draft?
                </h2>
                <p className='mt-3 text-sm leading-6 text-text-2'>
                    You have unsaved edits on this debate. Leaving now will discard them.
                </p>

                <div className='mt-6 flex flex-wrap justify-end gap-3'>
                    <Button type='button' variant='ghost' onClick={onStay}>
                        Stay and continue editing
                    </Button>
                    <Button type='button' variant='danger' onClick={onLeave}>
                        Leave and discard
                    </Button>
                </div>
            </div>
        </div>
    );
}
