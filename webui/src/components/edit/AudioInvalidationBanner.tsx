export interface AudioInvalidationBannerProps {
    visible: boolean;
}

export function AudioInvalidationBanner({ visible }: AudioInvalidationBannerProps) {
    if (!visible) {
        return null;
    }

    return (
        <div className='rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800'>
            Editing the topic or agent names will mark the current audio as outdated. You can still
            save now and regenerate audio later.
        </div>
    );
}
