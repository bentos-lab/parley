export type BadgeVariant =
    | 'default'
    | 'accent'
    | 'agent-1'
    | 'agent-2'
    | 'agent-3'
    | 'agent-4'
    | 'success'
    | 'danger';

export interface BadgeProps {
    label: string;
    variant?: BadgeVariant;
    className?: string;
}

const badgeVariantClasses: Record<BadgeVariant, string> = {
    default: 'border border-border bg-bg-elevated text-text-3',
    accent: 'border border-accent-dim bg-accent-bg text-accent',
    'agent-1': 'border border-agent-1 bg-agent-1-bg text-agent-1',
    'agent-2': 'border border-agent-2 bg-agent-2-bg text-agent-2',
    'agent-3': 'border border-agent-3 bg-agent-3-bg text-agent-3',
    'agent-4': 'border border-agent-4 bg-agent-4-bg text-agent-4',
    success: 'border border-agent-1 bg-agent-1-bg text-agent-1',
    danger: 'border border-agent-2 bg-agent-2-bg text-agent-2',
};

export function Badge({ label, variant = 'default', className = '' }: BadgeProps) {
    return (
        <span
            className={[
                'inline-flex items-center rounded-sm px-2 py-0.5 text-[10px] leading-none font-medium font-sans',
                badgeVariantClasses[variant],
                className,
            ].join(' ')}
        >
            {label}
        </span>
    );
}
