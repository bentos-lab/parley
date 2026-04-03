import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/Button';

export function EmptyState() {
    const navigate = useNavigate();

    return (
        <div className='flex flex-col items-center justify-center py-16 px-4 text-center'>
            <div className='w-20 h-20 rounded-full bg-bg-elevated border-2 border-border-mid flex items-center justify-center mb-6'>
                <svg
                    viewBox='0 0 24 24'
                    fill='none'
                    stroke='currentColor'
                    strokeWidth='1.5'
                    strokeLinecap='round'
                    strokeLinejoin='round'
                    className='w-9 h-9 text-text-3'
                    aria-hidden='true'
                >
                    <path d='M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z' />
                </svg>
            </div>

            <h2 className='text-lg font-display font-medium text-text-1 mb-2'>No debates yet</h2>
            <p className='text-text-2 text-sm mb-6 max-w-xs'>
                Create your first debate to get started with AI-powered argumentative exchange.
            </p>

            <Button variant='accent' onClick={() => navigate('/debates/new')}>
                Create First Debate
            </Button>
        </div>
    );
}
