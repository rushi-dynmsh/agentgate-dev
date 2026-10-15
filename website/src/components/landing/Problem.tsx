import React from 'react';
import { IconShieldOff, IconKey, IconEye } from './icons';

const ITEMS = [
  { icon: <IconShieldOff />, h: 'Agents are unpredictable', p: "You cannot guarantee a model won't hallucinate a destructive tool call or be steered by a prompt injection." },
  { icon: <IconKey />, h: 'Backend credentials are fragile', p: 'Handing an agent a database password or API key creates a large blast radius that nobody is watching.' },
  { icon: <IconEye />, h: 'Blind spots are unacceptable', p: 'When an agent acts you need a tamper-evident record of exactly what it did, not just its chat log.' },
];

export default function Problem() {
  return (
    <section id="problem" className="lp-section">
      <div className="lp-wrap">
        <div className="lp-head">
          <span className="lp-eyebrow">The problem</span>
          <h2 className="lp-title">Agents can act. Nothing is checking.</h2>
          <p className="lp-lead">Traditional firewalls and IAM see network traffic and logins. They cannot see what an agent is asking a tool to do.</p>
        </div>
        <div className="lp-grid">
          {ITEMS.map((i) => (
            <div className="lp-card" key={i.h}>
              <div className="lp-icon">{i.icon}</div>
              <h3>{i.h}</h3>
              <p>{i.p}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
