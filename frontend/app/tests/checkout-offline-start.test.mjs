import assert from 'node:assert/strict'
import { test } from 'node:test'

const logic = await import('../src/features/pos/logic.ts')
const { resolveCheckoutShiftGate } = logic

// Karakterisasi perbaikan bug checkout offline-penuh (Unit A):
// checkout yang DIMULAI saat perangkat sudah offline penuh harus
// langsung masuk antrean offline — bukan menggantung menunggu refetch
// status shift, dan bukan terkirim sebagai penjualan online saat
// koneksi pulih. Keputusan itu kini murni dari cache + status online
// di awal checkout lewat resolveCheckoutShiftGate (fitur pos/logic).

test('offline from the start with an active cached shift goes straight to the offline queue', () => {
  assert.equal(typeof resolveCheckoutShiftGate, 'function')
  // Tanpa refreshedHasShift sama sekali: keputusan offline tidak
  // pernah menunggu hasil refresh.
  assert.equal(
    resolveCheckoutShiftGate({ online: false, cachedHasShift: true }),
    'allow',
  )
})

test('offline from the start with unknown cached shift still goes to the offline queue', () => {
  // Cache belum pernah terisi (null) bukan alasan menggantung atau
  // mengirim online: backend menegakkan shift saat antrean tersinkron.
  assert.equal(
    resolveCheckoutShiftGate({ online: false, cachedHasShift: null }),
    'allow',
  )
})

test('offline from the start is blocked only when the cache proves there is no shift', () => {
  assert.equal(
    resolveCheckoutShiftGate({ online: false, cachedHasShift: false }),
    'block-no-shift',
  )
})

test('offline decision never asks for a refresh', () => {
  for (const cachedHasShift of [true, false, null]) {
    assert.notEqual(
      resolveCheckoutShiftGate({ online: false, cachedHasShift }),
      'needs-refresh',
      `cached=${String(cachedHasShift)} tidak boleh memicu refresh saat offline`,
    )
  }
})

test('online checkout still requires a shift refresh before proceeding', () => {
  assert.equal(
    resolveCheckoutShiftGate({ online: true, cachedHasShift: true }),
    'needs-refresh',
  )
  assert.equal(
    resolveCheckoutShiftGate({ online: true, cachedHasShift: null }),
    'needs-refresh',
  )
  // Bahkan cache yang menyatakan tidak ada shift tidak memblokir
  // jalur online sebelum refresh: status bisa saja baru berubah.
  assert.equal(
    resolveCheckoutShiftGate({ online: true, cachedHasShift: false }),
    'needs-refresh',
  )
})

test('online checkout is blocked only when the refreshed shift status is false', () => {
  assert.equal(
    resolveCheckoutShiftGate({ online: true, cachedHasShift: true, refreshedHasShift: false }),
    'block-no-shift',
  )
  assert.equal(
    resolveCheckoutShiftGate({ online: true, cachedHasShift: false, refreshedHasShift: true }),
    'allow',
  )
  // Status tak diketahui setelah refresh (gagal jaringan) tidak
  // memblokir — sama seperti perilaku lama yang hanya memblokir
  // pada nilai false; backend tetap penegak terakhir (409).
  assert.equal(
    resolveCheckoutShiftGate({ online: true, cachedHasShift: null, refreshedHasShift: null }),
    'allow',
  )
})
