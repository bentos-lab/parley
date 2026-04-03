import { ReactNode } from 'react';

export interface BrowseLayoutProps {
    children: ReactNode;
}

export function BrowseLayout({ children }: BrowseLayoutProps) {
    return (
        <div className='flex flex-1 flex-col items-center justify-center gap-8 overflow-auto p-8 sm:p-10'>
            {children}
        </div>
    );
}
