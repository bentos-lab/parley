export type AvatarShape = 'round' | 'square';
export type AvatarSize = 'sm' | 'md' | 'lg';
export type AvatarColor = 'agent-1' | 'agent-2' | 'agent-3' | 'agent-4' | 'agent-user' | 'default';

export interface AvatarProps {
    shape?: AvatarShape;
    size?: AvatarSize;
    src?: string;
    initials?: string;
    color?: AvatarColor;
    alt?: string;
    className?: string;
}

const sizeClasses: Record<AvatarSize, string> = {
    sm: 'h-5 w-5 text-[8px]',
    md: 'h-7 w-7 text-[10px]',
    lg: 'h-10 w-10 text-xs',
};

const colorClasses: Record<AvatarColor, string> = {
    default: 'border border-border-mid bg-bg-elevated text-text-2',
    'agent-1': 'border border-agent-1 bg-agent-1-bg text-agent-1',
    'agent-2': 'border border-agent-2 bg-agent-2-bg text-agent-2',
    'agent-3': 'border border-agent-3 bg-agent-3-bg text-agent-3',
    'agent-4': 'border border-agent-4 bg-agent-4-bg text-agent-4',
    'agent-user': 'border border-agent-user bg-agent-user-bg text-agent-user',
};

export function Avatar({
    shape = 'round',
    size = 'md',
    src,
    initials,
    color = 'default',
    alt = '',
    className = '',
}: AvatarProps) {
    const shapeClass = shape === 'round' ? 'rounded-full' : 'rounded-sm';

    if (src) {
        return (
            <img
                src={src}
                alt={alt}
                className={['shrink-0 object-cover', sizeClasses[size], shapeClass, className].join(
                    ' ',
                )}
            />
        );
    }

    return (
        <div
            aria-label={alt || initials || 'avatar'}
            className={[
                'inline-flex shrink-0 items-center justify-center font-mono font-medium',
                sizeClasses[size],
                shapeClass,
                colorClasses[color],
                className,
            ].join(' ')}
        >
            {initials ? initials.slice(0, 2).toUpperCase() : '?'}
        </div>
    );
}
