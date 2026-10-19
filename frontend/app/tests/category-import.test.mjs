import assert from 'node:assert/strict'
import { test } from 'node:test'

const { importCategoryFile } = await import('../src/features/catalog/components/category-import.ts')

test('category import stays busy while an Excel file is parsed and uploaded', async () => {
  let busy = false
  let finishParsing
  let finishUploading
  let startUploading
  let uploadedFile
  const uploadStarted = new Promise((resolve) => { startUploading = resolve })
  const parsing = new Promise((resolve) => { finishParsing = resolve })
  const uploading = new Promise((resolve) => { finishUploading = resolve })

  const operation = importCategoryFile(
    new File(['workbook'], 'categories.xlsx'),
    () => parsing,
    (file) => {
      uploadedFile = file
      startUploading()
      return uploading
    },
    (value) => { busy = value },
  )

  assert.equal(busy, true)
  finishParsing([['nama'], ['Alat Tulis']])
  await uploadStarted
  assert.equal(busy, true)
  assert.equal(uploadedFile.name, 'import.csv')
  assert.equal(await uploadedFile.text(), 'nama\nAlat Tulis')

  finishUploading()
  await operation
  assert.equal(busy, false)
})
