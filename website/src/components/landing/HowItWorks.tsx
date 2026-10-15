import React from 'react';
import styles from './HowItWorks.module.css';

const STEPS = [
  { h: 'Write the policy', p: 'Say what each role may do, by tool risk level. A no-code builder generates the Cedar policy for you.' },
  { h: 'Validate', p: 'The policy is checked for structure and safety before anything else happens, so mistakes surface immediately.' },
  { h: 'Dry run', p: 'Replay sample requests against the candidate. See which decisions would change without touching live behaviour.' },
  { h: 'Activate, or roll back', p: 'Go live with a recorded reason. If something looks wrong, one click restores the previous version.' },
];

export default function HowItWorks() {
  return (
    <section id="how-it-works" className="lp-section lp-section--alt">
      <div className="lp-wrap">
        <div className="lp-head">
          <span className="lp-eyebrow">How it works</span>
          <h2 className="lp-title">Four checkpoints before any rule goes live</h2>
          <p className="lp-lead">Changing what agents can do is a powerful operation, so AgentGate makes it deliberate and reversible.</p>
        </div>
        <ol className={styles.steps}>
          {STEPS.map((s, i) => (
            <li className={styles.step} key={s.h}>
              <span className={styles.num}>{i + 1}</span>
              <h3>{s.h}</h3>
              <p>{s.p}</p>
            </li>
          ))}
        </ol>
      </div>
    </section>
  );
}
