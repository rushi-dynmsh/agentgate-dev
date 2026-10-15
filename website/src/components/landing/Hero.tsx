import React from 'react';
import styles from './Hero.module.css';
import FlowAnimation from './FlowAnimation';

const FACTS = [
  { k: 'Fail-closed', v: 'Unknown means denied' },
  { k: 'Cedar policy', v: 'Versioned, dry-run, rollback' },
  { k: 'Hash-chained audit', v: 'Independently verifiable' },
  { k: '60-second credentials', v: 'Agents never hold backend keys' },
];

export default function Hero() {
  return (
    <header id="hero" className={styles.heroSection}>
      <div className="lp-wrap">
        <span className="lp-eyebrow">AI agent authorization gateway</span>
        <h1 id="hero-title" className={styles.title}>
          Every agent action,<br />
          <span className={styles.accent}>checked before it happens.</span>
        </h1>
        <p className={styles.tagline}>
          AgentGate is a checkpoint in the only path between an AI agent and its tools. It enforces policy on
          every call, mints short-lived credentials, and records each decision in a tamper-evident log.
        </p>
        <div className={styles.buttons}>
          <a className="button button--primary button--lg" href="/docs/quickstart">
            Get started
          </a>
          <a className="button button--secondary button--lg" href="#see-it-block">
            See it block an attack
          </a>
        </div>

        <div className={styles.animationCard}>
          <FlowAnimation />
        </div>

        <ul className={styles.facts}>
          {FACTS.map((f) => (
            <li key={f.k}>
              <strong>{f.k}</strong>
              <span>{f.v}</span>
            </li>
          ))}
        </ul>
      </div>
    </header>
  );
}
