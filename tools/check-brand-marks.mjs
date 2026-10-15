#!/usr/bin/env node
// Fails if any copy of the AgentGate mark drifts from the master geometry,
// or if a published copy differs from its brand/ source.
import fs from 'node:fs';

const GEOMETRY = [
  '<rect x="12" y="12" width="40" height="40" rx="9"',
  '<path d="M12 32h14M38 32h14"',
  'stroke-width="5"',
  '<circle cx="32" cy="32" r="7"',
];
const marks = [
  'brand/favicon.svg', 'brand/logo/mark.svg', 'brand/logo/mark-mono.svg', 'brand/logo/mark-dark.svg',
  'brand/logo/wordmark.svg', 'brand/logo/wordmark-dark.svg',
];
const copies = {
  'brand/favicon.svg': ['website/static/img/favicon.svg', 'frontend/app/public/favicon.svg'],
  'brand/logo/mark.svg': ['website/static/img/mark.svg'],
  'brand/logo/mark-dark.svg': ['website/static/img/mark-dark.svg'],
  'brand/logo/wordmark.svg': ['website/static/img/logo.svg'],
  'brand/logo/wordmark-dark.svg': ['website/static/img/logo-dark.svg'],
};
let bad = false;
for (const f of marks) {
  const s = fs.readFileSync(f, 'utf8');
  for (const g of GEOMETRY) if (!s.includes(g)) { console.error(`${f}: mark geometry drifted (missing ${g})`); bad = true; }
}
for (const [src, list] of Object.entries(copies)) for (const c of list) {
  if (fs.readFileSync(src, 'utf8') !== fs.readFileSync(c, 'utf8')) { console.error(`${c} differs from ${src}`); bad = true; }
}
if (bad) process.exit(1);
console.log('brand marks consistent');
