# React + TypeScript + Vite

## Verified

The UI has deterministic, backend-independent tests runnable with `npm test`:

- Policy tab filtering and policy lifecycle UI behavior.
- Activation remains pending until the governance client promise resolves; the UI does not report success early.
- Full mock policy lifecycle: list, validate, create candidate, dry-run, activate, and rollback.
- Existing AgentGate decision fixtures render ALLOW, DENY, transport-error, and stale states correctly.

Tests use Vitest, JSDOM, and React Testing Library. They do not require AgentGate or a live database.

The current checkout's real React application lives under `frontend/app/`. There is no separate
`admin-ui/` directory and no Decision Tester page; the decision-fixture suite tests the existing
contract fixtures directly rather than inventing a second screen. Audit and Tools remain fixture
surfaces until the backend read APIs are implemented and their contracts are reviewed.

This template provides a minimal setup to get React working in Vite with HMR and some Oxlint rules.

Currently, two official plugins are available:

- [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react) uses [Oxc](https://oxc.rs)
- [@vitejs/plugin-react-swc](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react-swc) uses [SWC](https://swc.rs/)

## React Compiler

The React Compiler is not enabled on this template because of its impact on dev & build performances. To add it, see [this documentation](https://react.dev/learn/react-compiler/installation).

## Expanding the Oxlint configuration

If you are developing a production application, we recommend enabling type-aware lint rules by installing `oxlint-tsgolint` and editing `.oxlintrc.json`:

```json
{
  "$schema": "./node_modules/oxlint/configuration_schema.json",
  "plugins": ["react", "typescript", "oxc"],
  "options": {
    "typeAware": true
  },
  "rules": {
    "react/rules-of-hooks": "error",
    "react/only-export-components": ["warn", { "allowConstantExport": true }]
  }
}
```

See the [Oxlint rules documentation](https://oxc.rs/docs/guide/usage/linter/rules) for the full list of rules and categories.
