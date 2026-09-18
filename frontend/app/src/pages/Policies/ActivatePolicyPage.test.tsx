import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { ActivatePolicyPage } from './ActivatePolicyPage';
import { useGovernance } from '../../state/GovernanceProvider';
import { useToast } from '../../state/ToastProvider';
import { governanceState, policy } from '../../test/test-fixtures';

const navigate = vi.hoisted(() => vi.fn());

vi.mock('../../state/GovernanceProvider', () => ({
  useGovernance: vi.fn(),
}));

vi.mock('../../state/ToastProvider', () => ({
  useToast: vi.fn(),
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: () => navigate };
});

const mockedUseGovernance = vi.mocked(useGovernance);
const mockedUseToast = vi.mocked(useToast);

describe('ActivatePolicyPage', () => {
  it('does not show success or navigate until the backend confirms activation', async () => {
    let resolveActivation!: (value: { workspace_id: string; active_version: string; activated_at: string }) => void;
    const activation = new Promise<{ workspace_id: string; active_version: string; activated_at: string }>((resolve) => {
      resolveActivation = resolve;
    });
    const activate = vi.fn(() => activation);

    mockedUseGovernance.mockReturnValue({
      state: governanceState([policy('active-v1', 'active'), policy('candidate-v2', 'candidate')]),
      ready: true,
      loadPolicies: vi.fn(),
      validateDraft: vi.fn(),
      submitCandidate: vi.fn(),
      activate,
      rollback: vi.fn(),
      dryRunCompare: vi.fn(),
    });
    mockedUseToast.mockReturnValue({ showToast: vi.fn() });
    render(
      <MemoryRouter initialEntries={['/policies/candidate-v2/activate']}>
        <Routes>
          <Route path="/policies/:version/activate" element={<ActivatePolicyPage />} />
        </Routes>
      </MemoryRouter>
    );

    fireEvent.click(screen.getByRole('checkbox'));
    fireEvent.click(screen.getByRole('button', { name: 'Activate policy' }));

    expect(activate).toHaveBeenCalledWith('candidate-v2');
    expect(screen.getByRole('button', { name: 'Activating…' })).toBeInTheDocument();
    expect(navigate).not.toHaveBeenCalled();

    resolveActivation({ workspace_id: 'default', active_version: 'candidate-v2', activated_at: '2026-09-18T12:00:00Z' });
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/policies'));
  });
});
