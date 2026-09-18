import { expect, it } from 'vitest';
import { MockGovernanceClient } from '@contract/api/governanceClient';
import { PolicyLifecycleStore } from '@contract/state/policyLifecycleState';

const POLICY = 'permit(principal, action, resource);';

it('completes the deterministic policy lifecycle against MockGovernanceClient', async () => {
  const client = new MockGovernanceClient();
  const store = new PolicyLifecycleStore(client, 'default');

  await store.loadPolicies();
  expect(store.getState().policies).toHaveLength(0);

  const validation = await store.validateDraft(POLICY);
  expect(validation.valid).toBe(true);
  expect(validation.version).toBeDefined();

  const candidate = await store.submitCandidate(POLICY, 'Test candidate policy');
  expect(candidate.state).toBe('candidate');
  expect(store.getState().policies).toHaveLength(1);

  const dryRun = await store.dryRunCompare(candidate.version, []);
  expect(dryRun.candidate_version).toBe(candidate.version);
  expect(store.getState().activeVersion).toBeUndefined();

  const activated = await store.activate(candidate.version);
  expect(activated.active_version).toBe(candidate.version);
  expect(store.getState().activeVersion).toBe(candidate.version);
  expect(store.getState().activationStatus).toBe('confirmed');

  const rollbackCandidate = await store.submitCandidate(
    'permit(principal, action, resource) when { true };',
    'Rollback target'
  );
  const rolledBack = await store.rollback(rollbackCandidate.version);
  expect(rolledBack.active_version).toBe(rollbackCandidate.version);
  expect(store.getState().activeVersion).toBe(rollbackCandidate.version);
  expect(store.getState().rollbackStatus).toBe('confirmed');
});
