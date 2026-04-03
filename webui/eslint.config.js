import js from '@eslint/js';
import tseslint from 'typescript-eslint';
import reactPlugin from 'eslint-plugin-react';
import reactHooksPlugin from 'eslint-plugin-react-hooks';
import reactRefreshPlugin from 'eslint-plugin-react-refresh';
import prettierConfig from 'eslint-config-prettier';
import prettierPlugin from 'eslint-plugin-prettier';

export default tseslint.config(
    // Global ignores
    {
        ignores: ['dist/**', 'node_modules/**', 'coverage/**', 'test-results/**'],
    },

    // Base JS recommended
    js.configs.recommended,

    // TypeScript recommended
    ...tseslint.configs.recommended,

    // React + hooks + refresh
    {
        files: ['**/*.{ts,tsx}'],
        plugins: {
            react: reactPlugin,
            'react-hooks': reactHooksPlugin,
            'react-refresh': reactRefreshPlugin,
            prettier: prettierPlugin,
        },
        languageOptions: {
            parserOptions: {
                ecmaFeatures: { jsx: true },
            },
        },
        settings: {
            react: { version: 'detect' },
        },
        rules: {
            // React
            ...reactPlugin.configs.recommended.rules,
            'react/react-in-jsx-scope': 'off', // Not needed with React 17+ JSX transform
            'react/prop-types': 'off', // TypeScript handles prop types

            // React hooks
            ...reactHooksPlugin.configs.recommended.rules,
            // The following patterns are intentional in this React Router v6 codebase:
            // - setState in effects: valid for syncing router navigation state to local state
            // - ref writes during render: valid React pattern for keeping callback refs fresh
            // - impure functions in useMemo/useCallback: e.g. Math.random() with [] deps is stable
            'react-hooks/set-state-in-effect': 'warn',
            'react-hooks/refs': 'warn',
            'react-hooks/purity': 'warn',

            // React refresh (Vite HMR)
            'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],

            // TypeScript
            '@typescript-eslint/no-unused-vars': [
                'error',
                { argsIgnorePattern: '^_', varsIgnorePattern: '^_' },
            ],
            '@typescript-eslint/no-explicit-any': 'warn',

            // Prettier — shows formatting issues as lint errors
            'prettier/prettier': 'error',
        },
    },

    // Disable ESLint formatting rules that conflict with Prettier
    prettierConfig,
);
