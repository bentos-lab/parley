import { StoreProvider } from 'easy-peasy';
import { RouterProvider } from 'react-router-dom';
import { router } from '@/app/router';
import { store } from '@/app/store';
import { RouteLoader } from '@/components/ui/RouteLoader';
import { ToastProvider } from '@/components/ui/ToastProvider';

export function AppProviders() {
    return (
        <StoreProvider store={store}>
            <ToastProvider>
                <RouterProvider router={router} fallbackElement={<RouteLoader />} />
            </ToastProvider>
        </StoreProvider>
    );
}
