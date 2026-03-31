---
name: e2e-test
description: >
  Run end-to-end tests and visual regression checks using Playwright. Use this skill when the
  user says "run e2e tests", "test the UI", "check for visual regressions", "QA the app",
  "smoke test", "run playwright", or wants to verify the full user flow works before deploying.
  Also use when the user reports a UI/UX bug and wants to write a regression test for it.
---

# E2E Testing

You are running end-to-end tests against the Metaloreian app using Playwright. These tests
verify the full user experience — from login to band discovery to playback — and catch visual
regressions between deploys.

## Prerequisites

- Working directory: `/home/imman/projects/metaloreian`
- Node.js installed
- Playwright installed: `cd frontend && npx playwright install`
- App running locally: backend on :8080, frontend on :5173
- For authenticated tests: a Spotify Premium account with valid credentials

## Setup (first time)

```bash
cd frontend
npm install -D @playwright/test
npx playwright install chromium
```

Add to `frontend/package.json` scripts:
```json
"e2e": "playwright test",
"e2e:ui": "playwright test --ui",
"e2e:headed": "playwright test --headed"
```

### Playwright Config

Create `frontend/playwright.config.ts`:
```typescript
import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  retries: 1,
  use: {
    baseURL: 'http://127.0.0.1:5173',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
  },
  projects: [
    { name: 'chromium', use: { browserName: 'chromium' } },
  ],
  webServer: {
    command: 'npm run dev',
    port: 5173,
    reuseExistingServer: true,
  },
});
```

## Test Structure

```
frontend/e2e/
├── login.spec.ts           # OAuth login flow
├── search.spec.ts          # Band search + navigation
├── band-page.spec.ts       # Band page rendering, similar bands
├── album-page.spec.ts      # Album page rendering, tracklist
├── player.spec.ts          # Playback controls
├── visual-regression.spec.ts  # Screenshot comparisons
└── fixtures/
    └── auth.ts             # Shared auth helper
```

## Writing Tests

### Functional E2E Tests

Test the critical user flows:

```typescript
// search.spec.ts
import { test, expect } from '@playwright/test';

test('search for a band and navigate to band page', async ({ page }) => {
  await page.goto('/');
  // Login would be needed here for full flow

  // Search
  await page.fill('.search-input', 'Metallica');
  await page.waitForSelector('.search-results');
  await page.click('.search-result-item >> text=Metallica');

  // Verify band page loaded
  await expect(page.locator('.band-name')).toContainText('Metallica');
  await expect(page.locator('.section-title')).toContainText('Discography');
});
```

### Visual Regression Tests

Capture screenshots and compare against baselines:

```typescript
// visual-regression.spec.ts
import { test, expect } from '@playwright/test';

test('login page visual', async ({ page }) => {
  await page.goto('/');
  await expect(page).toHaveScreenshot('login-page.png', {
    maxDiffPixelRatio: 0.01,
  });
});

test('band page visual', async ({ page }) => {
  // Navigate to a known band
  await page.goto('/band/125'); // Metallica
  await page.waitForSelector('.band-name');
  await expect(page).toHaveScreenshot('band-page-metallica.png', {
    maxDiffPixelRatio: 0.02,
  });
});
```

First run creates baseline screenshots in `e2e/*.png-snapshots/`. Subsequent runs compare
against them. Update baselines with `npx playwright test --update-snapshots`.

### Auth Helper

Spotify OAuth can't be fully automated (requires real login). Two approaches:

1. **Skip auth for unauthenticated pages** — login page, band pages (via direct URL if API works without auth)
2. **Storage state** — log in once manually, save cookies/storage, reuse:
```typescript
// fixtures/auth.ts
import { test as base } from '@playwright/test';

export const test = base.extend({
  storageState: 'frontend/e2e/.auth/user.json',
});
```
Generate with: `npx playwright codegen --save-storage=e2e/.auth/user.json http://127.0.0.1:5173`

## Running Tests

```bash
# All E2E tests
cd frontend && npx playwright test

# Specific test file
npx playwright test e2e/search.spec.ts

# With browser visible (debugging)
npx playwright test --headed

# Interactive UI mode
npx playwright test --ui

# Update visual snapshots after intentional UI changes
npx playwright test --update-snapshots

# Generate test from user actions
npx playwright codegen http://127.0.0.1:5173
```

## Writing Tests for UI/UX Bugs

When the user reports a bug:

1. **Reproduce** — write a test that demonstrates the broken behavior
2. **Assert the failure** — run the test, confirm it fails
3. **Fix the bug** — modify the code
4. **Assert the fix** — run the test, confirm it passes
5. **Commit both** — the fix and the regression test together

Example: "the player bar overlaps the band content on mobile"
```typescript
test('player bar does not overlap content on mobile', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 812 }); // iPhone
  await page.goto('/band/125');
  await page.waitForSelector('.band-page');

  const content = page.locator('.app-main');
  const player = page.locator('.player-bar');

  const contentBox = await content.boundingBox();
  const playerBox = await player.boundingBox();

  // Content should not extend into player bar area
  expect(contentBox!.y + contentBox!.height).toBeLessThanOrEqual(playerBox!.y);
});
```

## CI Integration

Add to `.github/workflows/deploy.yml` test job:

```yaml
- name: Install Playwright
  working-directory: frontend
  run: npx playwright install --with-deps chromium

- name: Run E2E tests
  working-directory: frontend
  run: npx playwright test
```

Visual regression tests in CI need consistent rendering — use the Docker-based approach:
```yaml
- name: Run E2E tests
  working-directory: frontend
  run: npx playwright test --project=chromium
```

## What to Test (Priority Order)

| Priority | Flow | Why |
|----------|------|-----|
| 1 | Login page renders | Entry point, first impression |
| 2 | Search returns results | Core functionality |
| 3 | Band page loads with all sections | Most viewed page |
| 4 | Album page loads with tracklist | Second most viewed |
| 5 | Similar bands section renders | New feature, beta |
| 6 | Player controls work | Requires Spotify Premium |
| 7 | Responsive layouts (mobile) | Common source of bugs |
| 8 | Visual regression on key pages | Catch CSS regressions |

## Visual Regression Workflow

After any CSS/UI change:
1. Run `npx playwright test` — if screenshots differ, tests fail
2. Review the diff: `npx playwright show-report`
3. If the change is intentional: `npx playwright test --update-snapshots`
4. If unintentional: fix the CSS regression
5. Commit updated snapshots with the code change

Snapshot files (`*.png`) should be committed to the repo so CI can compare against them.
