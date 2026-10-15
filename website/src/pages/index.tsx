import React from 'react';
import Layout from '@theme/Layout';
import Hero from '../components/landing/Hero';
import Problem from '../components/landing/Problem';
import HowItWorks from '../components/landing/HowItWorks';
import Trust from '../components/landing/Trust';
import AttackPlayer from '../components/landing/AttackPlayer';
import ConsoleSection from '../components/landing/ConsoleSection';
import QuickstartTeaser from '../components/landing/QuickstartTeaser';
import Head from '@docusaurus/Head';

import styles from './index.module.css';

export default function Home() {
  return (
    <Layout>
      <Head>
        <title>AgentGate | The AI Agent Control Plane</title>
        <meta name="description" content="AgentGate: A checkpoint in the only path between an AI agent and its tools." />
        <meta property="og:title" content="AgentGate | The AI Agent Control Plane" />
        <meta property="og:description" content="AgentGate: A checkpoint in the only path between an AI agent and its tools." />
        <meta property="og:image" content="/img/og.png" />
      </Head>
      <main className={styles.mainContainer}>
        <Hero />
        <Problem />
        <HowItWorks />
        
        <section id="see-it-block" className="lp-section">
          <div className="lp-wrap">
            <div className="lp-head">
              <span className="lp-eyebrow">See it in action</span>
              <h2 className="lp-title">Watch attacks get stopped</h2>
              <p className="lp-lead">Step through real scenarios: what the agent tried, what AgentGate checked, and the verdict.</p>
            </div>
            <AttackPlayer />
          </div>
        </section>
        <ConsoleSection />

        <Trust />
        <QuickstartTeaser />
      </main>
    </Layout>
  );
}
