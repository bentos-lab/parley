import { useRef, useState } from 'react';

type Tone = 'agent-1' | 'agent-2' | 'agent-3' | 'agent-4' | 'agent-user';

const toneClasses: Record<Tone, string> = {
    'agent-1': 'border-agent-1/30 bg-agent-1-bg/70 text-agent-1',
    'agent-2': 'border-agent-2/30 bg-agent-2-bg/70 text-agent-2',
    'agent-3': 'border-agent-3/30 bg-agent-3-bg/70 text-agent-3',
    'agent-4': 'border-agent-4/30 bg-agent-4-bg/70 text-agent-4',
    'agent-user': 'border-agent-user/30 bg-agent-user-bg/70 text-agent-user',
};

export interface RoundEditorProps {
    index: number;
    agentId: string;
    message: string;
    speakerLabel: string;
    tone: Tone;
    error?: string;
    defaultCollapsed?: boolean;
    onCommit: (message: string) => void;
}

function escapeHtml(value: string) {
    return value
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;');
}

function normalizeEditorText(value: string) {
    return value.replace(/\r\n/g, '\n').replace(/\u00a0/g, '');
}

export function RoundEditor({
    index,
    agentId,
    message,
    speakerLabel,
    tone,
    error,
    defaultCollapsed = true,
    onCommit,
}: RoundEditorProps) {
    const editorRef = useRef<HTMLDivElement>(null);
    const [initialHtml] = useState(() => escapeHtml(message).replace(/\n/g, '<br />'));
    const [isExpanded, setIsExpanded] = useState(!defaultCollapsed);

    return (
        <article className='rounded-2xl border border-border bg-bg-surface/80 shadow-sm backdrop-blur-sm overflow-hidden'>
            <input type='hidden' name={`round-${index}-agent-id`} value={agentId} />
            <input type='hidden' name={`round-${index}-message`} value={message} />

            <button
                type='button'
                onClick={() => setIsExpanded(!isExpanded)}
                className='w-full flex items-center justify-between gap-3 p-4 cursor-pointer hover:bg-bg-hover/50 transition-colors'
            >
                <div className='flex items-center gap-3'>
                    <span className='font-mono text-[10px] uppercase tracking-[0.18em] text-text-3'>
                        Round {index + 1}
                    </span>
                    <span
                        className={[
                            'inline-flex items-center rounded-full border px-2.5 py-1 text-[10px] font-semibold uppercase tracking-[0.14em]',
                            toneClasses[tone],
                        ].join(' ')}
                    >
                        {speakerLabel}
                    </span>
                </div>
                <svg
                    className={`h-4 w-4 text-text-3 transition-transform duration-200 ${
                        isExpanded ? 'rotate-180' : ''
                    }`}
                    fill='none'
                    viewBox='0 0 24 24'
                    stroke='currentColor'
                    strokeWidth={2}
                >
                    <path strokeLinecap='round' strokeLinejoin='round' d='M19 9l-7 7-7-7' />
                </svg>
            </button>

            <div
                className={`transition-all duration-200 ease-in-out overflow-hidden ${
                    isExpanded ? 'max-h-[2000px] opacity-100' : 'max-h-0 opacity-0'
                }`}
            >
                <div className='px-4 pb-4'>
                    <div
                        ref={editorRef}
                        className='min-h-28 rounded-xl border border-border bg-bg-base px-4 py-3 text-sm leading-6 text-text-1 outline-none transition focus:border-accent-dim focus:ring-1 focus:ring-accent/30'
                        contentEditable
                        suppressContentEditableWarning
                        role='textbox'
                        aria-label={`Round ${index + 1} message`}
                        data-testid={`round-editor-${index}`}
                        dangerouslySetInnerHTML={{ __html: initialHtml }}
                        onBlur={() =>
                            onCommit(normalizeEditorText(editorRef.current?.innerText ?? ''))
                        }
                    />

                    {error ? <p className='mt-2 text-xs text-error'>{error}</p> : null}
                </div>
            </div>
        </article>
    );
}
