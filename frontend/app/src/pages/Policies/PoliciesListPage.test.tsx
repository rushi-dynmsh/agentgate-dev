import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { PoliciesListPage } from './PoliciesListPage';
import { useGovernance } from '../../state/GovernanceProvider';
import { renderWithRouter, governanceState, policy } from '../../test/test-fixtures';

vi.mock('../../state/GovernanceProvider', () => ({
  useGovernance: vi.fn(),
}));

const mockedUseGovernance = vi.mocked(useGovernance);

describe('PoliciesListPage', () => {
  it('filters policy rows when All, Active, Candidates, and History are clicked', () => {
    mockedUseGovernance.mockReturnValue({
      state: governanceState([
        policy('active-v1', 'active'),
        policy('candidate-v2', 'candidate'),
        policy('history-v0', 'historical'),
      ]),
      ready: true,
      loadPolicies: vi.fn(),
      validateDraft: vi.fn(),
      submitCandidate: vi.fn(),
      activate: vi.fn(),
      rollback: vi.fn(),
      dryRunCompare: vi.fn(),
    });

    render(renderWithRouter(<PoliciesListPage />, ['/policies']));

    expect(screen.getByText('active policy')).toBeInTheDocument();
    expect(screen.getByText('candidate policy')).toBeInTheDocument();
    expect(screen.getByText('historical policy')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /Active 1/ }));
    expect(screen.getByText('active policy')).toBeInTheDocument();
    expect(screen.queryByText('candidate policy')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /Candidates 1/ }));
    expect(screen.getByText('candidate policy')).toBeInTheDocument();
    expect(screen.queryByText('active policy')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: /History 1/ }));
    expect(screen.getByText('historical policy')).toBeInTheDocument();
    expect(screen.queryByText('candidate policy')).not.toBeInTheDocument();
  });
});
