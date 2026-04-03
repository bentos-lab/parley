import { StoreProvider } from 'easy-peasy';
import { RouterProvider } from 'react-router-dom';
import { router } from '@/app/router';
import { store } from '@/app/store';
import { RouteLoader } from '@/components/ui/RouteLoader';

export function AppProviders() {
    return (
        <StoreProvider store={store}>
            <RouterProvider router={router} fallbackElement={<RouteLoader />} />
        </StoreProvider>
    );
}
