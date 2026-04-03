import { Input } from '@/components/ui/Input';

type AgentDraft = {
    id: string;
    name: string;
    stance: string;
};

type AgentField = 'name' | 'stance';

export interface AgentEditCardProps {
    index: number;
    agent: AgentDraft;
    errors: Record<string, string>;
    onChange: (field: AgentField, value: string) => void;
    onBlur: (field: AgentField) => void;
}

function getError(errors: Record<string, string>, index: number, field: AgentField) {
    return errors[`agents_${index}_${field}`];
}

export function AgentEditCard({ index, agent, errors, onChange, onBlur }: AgentEditCardProps) {
    return (
        <article className='rounded-2xl border border-border bg-bg-surface/80 p-4 shadow-sm backdrop-blur-sm'>
            <input type='hidden' name={`agent-${index}-id`} value={agent.id} />

            <div className='mb-4 flex items-center justify-between gap-3'>
                <div>
                    <p className='font-mono text-[10px] uppercase tracking-[0.18em] text-text-3'>
                        Agent {index + 1}
                    </p>
                    <p className='mt-1 text-sm text-text-2'>
                        Update speaker metadata without leaving the page.
                    </p>
                </div>
            </div>

            <div className='space-y-4'>
                <div>
                    <label
                        className='mb-1 block text-xs text-text-2'
                        htmlFor={`agent-${index}-name`}
                    >
                        Name
                    </label>
                    <Input
                        id={`agent-${index}-name`}
                        name={`agent-${index}-name`}
                        aria-label={`Agent ${index + 1} Name`}
                        value={agent.name}
                        onChange={(event) => onChange('name', event.target.value)}
                        onBlur={() => onBlur('name')}
                    />
                    {getError(errors, index, 'name') ? (
                        <p className='mt-1 text-xs text-error'>{getError(errors, index, 'name')}</p>
                    ) : null}
                </div>

                <div>
                    <label
                        className='mb-1 block text-xs text-text-2'
                        htmlFor={`agent-${index}-stance`}
                    >
                        Stance
                    </label>
                    <Input
                        id={`agent-${index}-stance`}
                        name={`agent-${index}-stance`}
                        aria-label={`Agent ${index + 1} Stance`}
                        value={agent.stance}
                        onChange={(event) => onChange('stance', event.target.value)}
                        onBlur={() => onBlur('stance')}
                    />
                    {getError(errors, index, 'stance') ? (
                        <p className='mt-1 text-xs text-error'>
                            {getError(errors, index, 'stance')}
                        </p>
                    ) : null}
                </div>
            </div>
        </article>
    );
}
