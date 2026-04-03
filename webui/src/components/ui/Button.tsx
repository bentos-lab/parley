import { type ButtonHTMLAttributes } from 'react';

export type ButtonVariant = 'accent' | 'ghost' | 'secondary' | 'danger';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: ButtonVariant;
}

const variantClasses: Record<ButtonVariant, string> = {
    accent: 'bg-accent text-bg-base border border-accent font-medium hover:opacity-90',
    ghost: 'bg-transparent text-accent border border-border-mid hover:bg-bg-elevated',
    secondary: 'bg-bg-elevated text-text-1 border border-border hover:bg-bg-hover',
    danger: 'bg-agent-2-bg text-agent-2 border border-agent-2 hover:opacity-90',
};

export function Button({
    variant = 'secondary',
    className = '',
    disabled,
    children,
    ...props
}: ButtonProps) {
    return (
        <button
            className={[
                'inline-flex items-center justify-center gap-1.5 rounded px-3 py-1.5 text-xs font-sans transition',
                variantClasses[variant],
                disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer',
                className,
            ].join(' ')}
            disabled={disabled}
            {...props}
        >
            {children}
        </button>
    );
}
