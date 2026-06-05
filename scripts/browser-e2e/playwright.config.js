const { defineConfig, devices } = require('@playwright/test')

module.exports = defineConfig({
  testDir: '.',
  testMatch: ['e2e.spec.js'],
  fullyParallel: false,
  workers: 1,
  reporter: [['list'], ['html', { open: 'never', outputFolder: '/tmp/playwright-e2e/report' }]],
  use: {
    headless: true,
    ignoreHTTPSErrors: true,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
})
