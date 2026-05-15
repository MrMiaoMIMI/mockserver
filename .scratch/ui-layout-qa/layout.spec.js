const { test, expect } = require('playwright/test')

const rulesUrl = 'http://localhost:6173/rulesets/http-default-test-1-f44939f1/rules?workbench=editor&rule=asdsadsdwq'

test('rule action editors use wide payload surfaces with examples', async ({ page }) => {
  const errors = []
  page.on('pageerror', (error) => errors.push(error.message))
  page.on('console', (message) => {
    if (message.type() === 'error') errors.push(message.text())
  })

  await page.setViewportSize({ width: 1440, height: 980 })
  await page.goto(rulesUrl)
  await page.locator('.section-switch button', { hasText: 'Action' }).click()

  const bodyEditor = page.locator('.json-field', { hasText: 'body JSON' }).locator('.json-editor')
  await expect(bodyEditor).toBeVisible()
  const editorShellBox = await page.locator('.editor-shell').boundingBox()
  const bodyBox = await bodyEditor.boundingBox()
  expect(bodyBox.width / editorShellBox.width).toBeGreaterThan(0.75)
  expect(bodyBox.width).toBeGreaterThan(560)

  await page.locator('.action-type-grid button', { hasText: 'Sequence' }).click()
  const sequenceEditor = page.locator('.sequence-step').first().locator('.json-field', { hasText: 'body JSON' }).locator('.json-editor')
  await expect(sequenceEditor).toBeVisible()
  const sequenceBox = await sequenceEditor.boundingBox()
  expect(sequenceBox.width).toBeGreaterThan(560)

  await page.locator('.action-type-grid button', { hasText: 'Template' }).click()
  const templateTextarea = page.locator('.action-payload-item textarea')
  await expect(page.locator('.action-payload-item .json-editor', { hasText: 'Response Payload Template' })).toBeVisible()
  await expect(templateTextarea).toHaveAttribute('placeholder', /query \. "q1"/)

  await page.locator('.action-type-grid button', { hasText: 'CEL' }).click()
  const celTextarea = page.locator('.action-payload-item textarea')
  await expect(page.locator('.action-payload-item .json-editor', { hasText: 'Response Payload CEL' })).toBeVisible()
  await expect(celTextarea).toHaveAttribute('placeholder', /request\.path/)

  await page.screenshot({ path: '.scratch/ui-layout-qa/rule-action-layout.png', fullPage: true })
  expect(errors).toEqual([])
})

test('namespace editor dialog has a wider desktop workspace', async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 980 })
  await page.goto('http://localhost:6173/namespaces')
  await page.getByRole('button', { name: 'Create namespace' }).click()

  const dialog = page.locator('.el-dialog')
  await expect(dialog).toBeVisible()
  const dialogBox = await dialog.boundingBox()
  expect(dialogBox.width).toBeGreaterThan(1300)
  await page.screenshot({ path: '.scratch/ui-layout-qa/namespace-dialog-layout.png', fullPage: true })
})
