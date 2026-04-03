export function SkeletonCard() {
    return (
        <div className='p-4 border border-border rounded-md bg-bg-elevated'>
            {/* Title skeleton */}
            <div className='h-5 bg-bg-hover rounded w-2/3 mb-3 motion-safe:animate-pulse' />

            {/* Topic skeleton (2 lines) */}
            <div className='space-y-2 mb-3'>
                <div className='h-4 bg-bg-hover rounded w-4/5 motion-safe:animate-pulse' />
                <div className='h-4 bg-bg-hover rounded w-3/5 motion-safe:animate-pulse' />
            </div>

            {/* Metadata skeleton */}
            <div className='flex gap-2 flex-wrap'>
                <div className='h-4 bg-bg-hover rounded w-20 motion-safe:animate-pulse' />
                <div className='h-4 bg-bg-hover rounded w-16 motion-safe:animate-pulse' />
                <div className='h-4 bg-bg-hover rounded w-16 motion-safe:animate-pulse' />
            </div>
        </div>
    );
}

export interface SkeletonGridProps {
    count?: number;
}

/**
 * Grid of skeleton cards for the loading state.
 * Renders `count` cards (default 5) to fill the viewport during fetch.
 */
export function SkeletonGrid({ count = 5 }: SkeletonGridProps) {
    return (
        <div className='space-y-2'>
            {Array.from({ length: count }).map((_, i) => (
                <SkeletonCard key={i} />
            ))}
        </div>
    );
}
