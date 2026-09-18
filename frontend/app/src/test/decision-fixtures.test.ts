import { describe, expect, it } from 'vitest';
import { wireFixtures, uiFixtures } from '@contract/fixtures';
import { parseAuthorizationResult, parseTransportError } from '@contract/parsing/parse-authorization-result';
import { renderOperationState, toStaleView } from '@contract/view/decision-view';

function stateFromFixture(fixture: unknown) {
  const parsed = parseAuthorizationResult(fixture);
  return parsed.ok
    ? { status: 'succeeded' as const, data: parsed.value, confirmedAt: '2026-09-18T00:00:00Z' }
    : { status: 'apiError' as const, error: parsed.error, occurredAt: '2026-09-18T00:00:00Z' };
}

function toneFromState(state: ReturnType<typeof stateFromFixture>) {
  const view = renderOperationState(state);
  if (!('tone' in view)) throw new Error(`expected a rendered decision state, got ${view.status}`);
  return view.tone;
}

describe('decision fixtures used by the operator-facing decision surface', () => {
  it('renders the allow fixture as allow and every deny fixture as deny', () => {
    expect(toneFromState(stateFromFixture(wireFixtures.allow))).toBe('allow');
    for (const fixture of [
      wireFixtures.denyExplicitForbid,
      wireFixtures.denyNoMatchingPolicy,
      wireFixtures.denyInvalidIdentity,
      wireFixtures.denyUnknownTool,
      wireFixtures.denyEvaluationError,
      wireFixtures.denyMalformedRequest,
      wireFixtures.denyNoPolicyLoaded,
    ]) {
      expect(toneFromState(stateFromFixture(fixture))).toBe('deny');
    }
  });

  it('never turns transport or stale fixtures into an authorization decision', () => {
    const transportError = parseTransportError(wireFixtures.transportError, 400);
    expect(toneFromState({ status: 'apiError', error: transportError, occurredAt: 'now' })).toBe('error');
    const stale = toStaleView(uiFixtures.staleNoPriorData.reason, null);
    expect(stale.tone).toBe('stale');
    expect(stale.lastKnown).toBeUndefined();
  });
});
