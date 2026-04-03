import { forwardRef, type InputHTMLAttributes } from 'react';

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
    className?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
    { className = '', ...props },
    ref,
) {
    return (
        <input
            ref={ref}
            className={[
                'w-full rounded border border-border bg-bg-surface px-3 py-1.5 text-xs font-sans text-text-1',
                'placeholder:text-text-3 focus:border-accent-dim focus:outline-none focus:ring-1 focus:ring-accent/30',
                'disabled:cursor-not-allowed disabled:opacity-50',
                className,
            ].join(' ')}
            {...props}
        />
    );
});

Input.displayName = 'Input';
