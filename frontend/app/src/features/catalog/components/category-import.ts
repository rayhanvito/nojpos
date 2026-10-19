export async function importCategoryFile(
  file: File,
  parseImportFile: (file: File) => Promise<string[][]>,
  upload: (file: File) => Promise<void>,
  setBusy: (busy: boolean) => void,
): Promise<void> {
  setBusy(true)
  try {
    let importFile = file
    if (/\.(xlsx|xls)$/i.test(file.name)) {
      const rows = await parseImportFile(file)
      const csvText = rows.map((row) => row.map((cell) => (/[",\n]/.test(cell) ? '"' + cell.replace(/"/g, '""') + '"' : cell)).join(',')).join('\n')
      importFile = new File([csvText], 'import.csv', { type: 'text/csv' })
    }
    await upload(importFile)
  } finally {
    setBusy(false)
  }
}
