import ReactMarkdown from 'react-markdown';
import remarkBreaks from 'remark-breaks';
import remarkGfm from 'remark-gfm';

export interface MarkdownProps {
    content: string;
    streaming?: boolean;
}

function processContent(text: string) {
    return text.replace(/<pause\d+>/g, '\n\n').trim();
}

export function Markdown({ content, streaming = false }: MarkdownProps) {
    const processedContent = processContent(content);

    return (
        <div className='min-w-0 text-left text-[13px] leading-[1.7] text-text-1 break-words'>
            <ReactMarkdown
                remarkPlugins={[remarkGfm, remarkBreaks]}
                components={{
                    p: ({ children, ...props }) => (
                        <p className='mb-4 last:mb-0 whitespace-pre-wrap' {...props}>
                            {children}
                        </p>
                    ),
                    ul: ({ children, ...props }) => (
                        <ul className='mb-3 list-disc pl-5 last:mb-0' {...props}>
                            {children}
                        </ul>
                    ),
                    ol: ({ children, ...props }) => (
                        <ol className='mb-3 list-decimal pl-5 last:mb-0' {...props}>
                            {children}
                        </ol>
                    ),
                    li: ({ children, ...props }) => (
                        <li className='mb-1 last:mb-0' {...props}>
                            {children}
                        </li>
                    ),
                    a: ({ children, ...props }) => (
                        <a
                            className='text-accent underline underline-offset-2 hover:text-accent-hi'
                            {...props}
                        >
                            {children}
                        </a>
                    ),
                    strong: ({ children, ...props }) => (
                        <strong className='font-semibold text-text-1' {...props}>
                            {children}
                        </strong>
                    ),
                    em: ({ children, ...props }) => (
                        <em className='italic text-text-2' {...props}>
                            {children}
                        </em>
                    ),
                    blockquote: ({ children, ...props }) => (
                        <blockquote
                            className='mb-3 border-l-2 border-border-mid pl-3 italic text-text-2 last:mb-0'
                            {...props}
                        >
                            {children}
                        </blockquote>
                    ),
                    code: ({ children, className, ...props }) => {
                        const isInline = !className;
                        if (isInline) {
                            return (
                                <code
                                    className='rounded bg-bg-hover px-1.5 py-0.5 font-mono text-[12px] text-text-1'
                                    {...props}
                                >
                                    {children}
                                </code>
                            );
                        }

                        return (
                            <code className='font-mono text-[12px] text-text-1' {...props}>
                                {children}
                            </code>
                        );
                    },
                    pre: ({ children, ...props }) => (
                        <pre
                            className='mb-3 overflow-x-auto rounded-lg border border-border bg-bg-hover px-3 py-2 font-mono text-[12px] last:mb-0'
                            {...props}
                        >
                            {children}
                        </pre>
                    ),
                    hr: (props) => <hr className='my-4 border-border' {...props} />,
                    br: () => <br className='block h-2 content-[""]' />,
                    table: ({ children, ...props }) => (
                        <div className='mb-3 overflow-x-auto last:mb-0'>
                            <table
                                className='min-w-full border-collapse text-left text-[12px]'
                                {...props}
                            >
                                {children}
                            </table>
                        </div>
                    ),
                    thead: ({ children, ...props }) => (
                        <thead className='bg-bg-hover text-text-1' {...props}>
                            {children}
                        </thead>
                    ),
                    tbody: ({ children, ...props }) => (
                        <tbody className='divide-y divide-border-mid' {...props}>
                            {children}
                        </tbody>
                    ),
                    th: ({ children, ...props }) => (
                        <th className='px-3 py-2 font-semibold' {...props}>
                            {children}
                        </th>
                    ),
                    td: ({ children, ...props }) => (
                        <td className='px-3 py-2 text-text-2' {...props}>
                            {children}
                        </td>
                    ),
                }}
            >
                {processedContent}
            </ReactMarkdown>
            {streaming ? (
                <span className='ml-0.5 inline-block animate-[blink_0.8s_steps(1)_infinite] text-accent'>
                    ▌
                </span>
            ) : null}
        </div>
    );
}
