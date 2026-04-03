export type PillVariant = 'default' | 'accent' | 'muted';

export interface PillProps {
    label: string;
    variant?: PillVariant;
    className?: string;
}

const pillVariantClasses: Record<PillVariant, string> = {
    default: 'bg-bg-elevated text-text-2',
    accent: 'bg-accent-bg text-accent',
    muted: 'bg-bg-elevated text-text-3',
};

export function Pill({ label, variant = 'default', className = '' }: PillProps) {
    return (
        <span
            className={[
                'inline-flex items-center rounded px-1.5 py-0.5 text-[9px] leading-none font-mono',
                pillVariantClasses[variant],
                className,
            ].join(' ')}
        >
            {label}
        </span>
    );
}
