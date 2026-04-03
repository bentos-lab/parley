import { Sidebar } from '@/components/layout/Sidebar';
import { Topbar } from '@/components/layout/Topbar';
import { Outlet, useLoaderData, useMatches } from 'react-router-dom';
import { listDebates } from '@/services/api/debates';
import { DebateSummaryListSchema } from '@/services/api/schemas';
import type { Debate, DebateSummary } from '@/types';

// eslint-disable-next-line react-refresh/only-export-components
export async function loader() {
    const data = await listDebates();
    return DebateSummaryListSchema.parse(data);
}

export function RootLayout() {
    const debates = useLoaderData() as DebateSummary[];
    const matches = useMatches();

    // Find the deepest match that has a full debate object (from detail/edit/audio loaders)
    const debateMatch = [...matches].reverse().find((m) => {
        if (!m.data || typeof m.data !== 'object') return false;
        const data = m.data as Record<string, unknown>;
        return 'id' in data && 'normalizedName' in data && 'topic' in data;
    });

    const currentDebate = debateMatch?.data as Debate | undefined;

    return (
        <div className='flex h-screen w-screen overflow-hidden bg-bg-base'>
            <Sidebar debates={debates} />
            <div className='flex flex-1 flex-col overflow-hidden'>
                <Topbar debate={currentDebate ?? null} />
                <main className='flex-1 overflow-auto bg-bg-base'>
                    <Outlet />
                </main>
            </div>
        </div>
    );
}
