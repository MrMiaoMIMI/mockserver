const { test, expect } = require('playwright/test')

test('SPEX rules workspace can be resized and Rules tab header is compact', async ({ page }) => {
  const errors = []
  page.on('pageerror', (error) => errors.push(error.message))
  page.on('console', (message) => {
    if (message.type() === 'error') errors.push(message.text())
  })

  await page.setViewportSize({ width: 1440, height: 980 })
  await page.addInitScript(() => {
    window.localStorage.removeItem('mockserver.rules.workspace.split')
  })
  await page.goto('http://localhost:6173/rulesets/spex-default-spex-test-1-31be3183/rules?workbench=inspect&rule=12321')

  const header = page.locator('.workbench-header')
  await expectIntentHeader(page, 'Rules', 'Inspect, create, and edit rules')
  await page.locator('.workbench-header .task-rail button', { hasText: 'Test' }).click()
  await expectIntentHeader(page, 'Test', 'Simulate requests and inspect results')
  await page.locator('.workbench-header .task-rail button', { hasText: 'Release' }).click()
  await expectIntentHeader(page, 'Release', 'Validate, publish, and rollback')
  await page.locator('.workbench-header .task-rail button', { hasText: 'Rules' }).click()
  await expect(page.locator('.overview-hero')).toContainText('selected rule')

  const workspace = page.locator('.rules-workspace')
  const resizer = page.locator('.workspace-resizer')
  const manager = page.locator('.rules-main')
  await expect(resizer).toBeVisible()

  const workspaceBox = await workspace.boundingBox()
  const resizerBox = await resizer.boundingBox()
  const beforeBox = await manager.boundingBox()
  await page.mouse.move(resizerBox.x + resizerBox.width / 2, resizerBox.y + resizerBox.height / 2)
  await page.mouse.down()
  await page.mouse.move(workspaceBox.x + workspaceBox.width * 0.6, resizerBox.y + resizerBox.height / 2, { steps: 8 })
  await page.mouse.up()

  const afterBox = await manager.boundingBox()
  expect(afterBox.width).toBeGreaterThan(beforeBox.width + 120)
  await expect(resizer).toHaveAttribute('aria-valuenow', /[5-6][0-9]/)

  const savedSplit = await page.evaluate(() => window.localStorage.getItem('mockserver.rules.workspace.split'))
  expect(Number(savedSplit)).toBeGreaterThanOrEqual(55)

  await page.screenshot({ path: '.scratch/ui-layout-qa/spex-workspace-split.png', fullPage: true })
  expect(errors).toEqual([])
})

async function expectIntentHeader(page, title, detail) {
  const header = page.locator('.workbench-header')
  await expect(header).toContainText('Rule Workspace')
  await expect(header).toContainText(title)
  await expect(header).toContainText(detail)
  const tabButtons = header.locator('.task-rail button')
  await expect(tabButtons).toHaveCount(3)
  const tabBoxes = await tabButtons.evaluateAll((buttons) => buttons.map((button) => button.getBoundingClientRect()))
  expect(Math.max(...tabBoxes.map((box) => box.y)) - Math.min(...tabBoxes.map((box) => box.y))).toBeLessThan(4)
}
