import { useEffect, useRef, useState } from 'react';
import { Sidebar } from '@/components/layout/Sidebar';
import { Topbar } from '@/components/layout/Topbar';
import { useToast } from '@/components/ui/ToastProvider';
import { Outlet, useLoaderData, useMatches } from 'react-router-dom';
import { listDebates } from '@/services/api/debates';
import { getErrorMessage } from '@/services/api/http';
import { DebateSummaryListSchema } from '@/services/api/schemas';
import type { Debate, DebateSummary } from '@/types';

export interface RootLoaderData {
    debates: DebateSummary[];
    sidebarError: string | null;
}

// eslint-disable-next-line react-refresh/only-export-components
export async function loader(): Promise<RootLoaderData> {
    try {
        const data = await listDebates();
        return {
            debates: DebateSummaryListSchema.parse(data),
            sidebarError: null,
        };
    } catch (error) {
        return {
            debates: [],
            sidebarError: getErrorMessage(error, 'Cannot reach backend'),
        };
    }
}

const SIDEBAR_COLLAPSED_KEY = 'parley:sidebar-collapsed';

export function RootLayout() {
    const { debates, sidebarError } = useLoaderData() as RootLoaderData;
    const matches = useMatches();
    const toast = useToast();
    const shownToastRef = useRef(false);

    const [sidebarCollapsed, setSidebarCollapsed] = useState(() => {
        try {
            return localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === 'true';
        } catch {
            return false;
        }
    });

    const handleSidebarToggle = () => {
        setSidebarCollapsed((prev) => {
            const next = !prev;
            try {
                localStorage.setItem(SIDEBAR_COLLAPSED_KEY, String(next));
            } catch {
                // Ignore storage errors
            }
            return next;
        });
    };

    // Show toast once on mount if there's a sidebar error
    useEffect(() => {
        if (sidebarError && !shownToastRef.current) {
            shownToastRef.current = true;
            toast.error(sidebarError, { title: 'Connection issue' });
        }
    }, [sidebarError, toast]);

    // Find the deepest match that has a full debate object (from detail/edit/audio loaders)
    const debateMatch = [...matches].reverse().find((m) => {
        if (!m.data || typeof m.data !== 'object') return false;
        const data = m.data as Record<string, unknown>;
        return 'id' in data && 'normalizedName' in data && 'topic' in data;
    });

    const currentDebate = debateMatch?.data as Debate | undefined;

    return (
        <div className='flex h-screen w-screen overflow-hidden bg-bg-base'>
            <Sidebar
                debates={debates}
                sidebarError={sidebarError}
                collapsed={sidebarCollapsed}
                onToggle={handleSidebarToggle}
            />
            <div className='flex flex-1 flex-col overflow-hidden'>
                <Topbar debate={currentDebate ?? null} />
                <main className='flex-1 overflow-auto bg-bg-base'>
                    <Outlet />
                </main>
            </div>
        </div>
    );
}
