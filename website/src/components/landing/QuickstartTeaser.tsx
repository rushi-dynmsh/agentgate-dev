import React from 'react';
import styles from './QuickstartTeaser.module.css';

const COMMANDS = `git clone https://github.com/rushi-dynmsh/agentgate-dev.git
cd agentgate-dev
.\\deploy\\prod\\scripts\\demo-mode.ps1 -Reset`;

export default function QuickstartTeaser() {
  return (
    <section id="quickstart" className="lp-section lp-section--alt">
      <div className="lp-wrap">
        <div className={styles.panel}>
          <div className={styles.copy}>
            <span className="lp-eyebrow">Quickstart</span>
            <h2 className="lp-title">Run it on your machine</h2>
            <p className="lp-lead">
              One command starts the gateway, AgentGate, the identity provider, a sample backend and the admin console.
              Then connect a real MCP client and watch a call get blocked.
            </p>
            <div className={styles.action}>
              <a className="button button--primary button--lg" href="/docs/quickstart">Read the Quickstart</a>
              <a className="button button--secondary button--lg" href="/docs/developer/getting-started/overview">Contributor guide</a>
            </div>
          </div>
          <div className={styles.terminal}>
            <div className={styles.dots} aria-hidden="true"><i /><i /><i /></div>
            <pre><code>{COMMANDS}</code></pre>
          </div>
        </div>
      </div>
    </section>
  );
}
