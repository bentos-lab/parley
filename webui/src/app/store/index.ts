import { createStore } from 'easy-peasy';
import { injections } from './injections';
import { model, type StoreModel } from './model';

export const store = createStore<StoreModel>(model, {
    injections,
});

export type { StoreModel };
