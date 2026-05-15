const { test, expect } = require('playwright/test')

test('SPEX rule editor renders protocol response fields', async ({ page }) => {
  const errors = []
  page.on('pageerror', (error) => errors.push(error.message))
  page.on('console', (message) => {
    if (message.type() === 'error') errors.push(message.text())
  })

  await page.goto('http://127.0.0.1:6174/rulesets/spex-default-spex-test-1-31be3183/rules?workbench=editor&rule=12321')
  await page.locator('.section-switch button', { hasText: 'Action' }).click()

  await expect(page.getByText('code *')).toBeVisible()
  await expect(page.getByText('resp *')).toBeVisible()
  await expect(page.getByText('resp JSON')).toBeVisible()
  await expect(page.getByText('Response Payload JSON')).toHaveCount(0)
  await page.getByText('resp JSON').scrollIntoViewIfNeeded()
  await page.screenshot({ path: '.scratch/protocol-response-editor/qa/spex-action-editor.png', fullPage: true })
  expect(errors).toEqual([])
})
