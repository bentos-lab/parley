import type { RouteObject } from 'react-router-dom';
import { RootLayout, loader as rootLoader } from '@/routes/root/route';
import { RootErrorBoundary } from '@/routes/root/ErrorBoundary';
import { RouteLoader } from '@/components/ui/RouteLoader';

export const routes: RouteObject[] = [
    {
        path: '/',
        element: <RootLayout />,
        loader: rootLoader,
        errorElement: <RootErrorBoundary />,
        hydrateFallbackElement: <RouteLoader />,
        children: [
            {
                index: true,
                lazy: () =>
                    import('@/routes/debates/list/route').then((module) => ({
                        loader: module.loader,
                        Component: module.Component,
                        ErrorBoundary: module.ErrorBoundary,
                    })),
            },
            {
                path: 'debates',
                lazy: () => import('@/routes/debates/list/route'),
            },
            {
                path: 'debates/new',
                lazy: () => import('@/routes/debates/create/route'),
            },
            {
                path: 'debates/:debateId',
                lazy: () => import('@/routes/debates/detail/route'),
            },
            {
                path: 'debates/:debateId/edit',
                lazy: () => import('@/routes/debates/edit/route'),
            },
            {
                path: 'debates/:debateId/audio',
                lazy: () => import('@/routes/debates/audio/route'),
            },
        ],
    },
];
