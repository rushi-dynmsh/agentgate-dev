import React from 'react';
import { IconLock, IconLink, IconClock } from './icons';

export default function Trust() {
  return (
    <section id="trust" className="lp-section">
      <div className="lp-wrap">
        <div className="lp-head">
          <span className="lp-eyebrow ag-trust-badge">Zero trust</span>
          <h2 className="lp-title">Designed to fail safe</h2>
          <p className="lp-lead">Three guarantees that hold even when something else in the system breaks.</p>
        </div>
        <div className="lp-grid">
          <div className="lp-card">
            <div className="lp-icon"><IconLock /></div>
            <h3>Fails closed</h3>
            <p>A network timeout, malformed request or engine error all end the same way: the action is denied. <a href="/docs/security-and-limitations">Read the security model</a>.</p>
          </div>
          <div className="lp-card">
            <div className="lp-icon"><IconLink /></div>
            <h3>Tamper-evident audit</h3>
            <p>Every decision is written to a verifiable hash chain before the action runs. <a href="/docs/audit">Learn about the audit trail</a>.</p>
          </div>
          <div className="lp-card">
            <div className="lp-icon"><IconClock /></div>
            <h3>No shared credentials</h3>
            <p>Agents present short-lived tokens and never see backend passwords or API keys. <a href="/docs/credentials">See how credentials work</a>.</p>
          </div>
        </div>
      </div>
    </section>
  );
}
