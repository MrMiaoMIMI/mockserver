const { test, expect } = require('playwright/test')

test('SPEX simulation shows failed candidate rules as missed in Rule Manager', async ({ page }) => {
  const errors = []
  page.on('pageerror', (error) => errors.push(error.message))
  page.on('console', (message) => {
    if (message.type() === 'error') errors.push(message.text())
  })

  await page.setViewportSize({ width: 1440, height: 980 })
  await page.goto('http://localhost:6173/rulesets/spex-default-spex-test-1-31be3183/rules?workbench=simulate&rule=12321')

  await expect(page.locator('.rule-card').first()).toBeVisible()
  const eventJson = JSON.stringify({
    protocol: 'spex',
    namespace: 'default',
    request: {
      cmd: 'active.demo',
      req: {
        id: 'demo',
      },
      param: 'default',
    },
  }, null, 2)

  await page.locator('.workbench-panel:visible .json-editor textarea').fill(eventJson)
  await page.locator('.workbench-panel:visible button.wide-action', { hasText: 'Simulate' }).click()
  await expect(page.locator('.simulate-result-panel')).toBeVisible()

  await expect(page.locator('.rule-card', { hasText: '12321' })).toContainText('missed')
  await expect(page.locator('.rule-card', { hasText: '12321' })).not.toContainText('candidate')
  await expect(page.locator('.rule-card', { hasText: 'asxsadsa' })).toContainText('missed')
  await expect(page.locator('.rule-card', { hasText: 'asxsadsa' })).not.toContainText('candidate')
  expect(errors).toEqual([])
})
