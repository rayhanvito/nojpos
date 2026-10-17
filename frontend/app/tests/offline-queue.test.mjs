import assert from 'node:assert/strict'
import { test } from 'node:test'
import {
  createOfflineQueueItem,
  createClientTxId,
  isNetworkFailure,
  markQueueItemFailed,
  markQueueItemSynced,
  retryOfflineQueueItem,
  queueItemsForStore,
} from '../src/shared/offline-queue.ts'

const payload = {
  items: [{ productId: 'kopi-1', qty: 2 }],
  discount: 500,
  method: 'Cash',
  paid: 19500,
  clientTxId: '0e51d5a9-7054-47d5-a78b-eec3de4ab3a2',
}
const summary = { items: [{ name: 'Kopi', qty: 2 }], estimatedTotal: 19000 }
const storeId = 'store-17'
const createdAt = '2026-10-15T09:00:00.000Z'
const queueItem = (at = createdAt) => createOfflineQueueItem(payload, summary, storeId, at)

test('client transaction ID is a UUID suitable for the idempotency contract', () => {
  assert.match(createClientTxId(), /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i)
})

test('network failures remain retryable while HTTP errors need action', () => {
  assert.equal(isNetworkFailure(new TypeError('Failed to fetch'), true), true)
  assert.equal(isNetworkFailure({ status: 422, message: 'Validasi gagal' }, false), false)
  assert.equal(isNetworkFailure(new Error('Offline'), false), true)
})

test('queue creation keeps the complete checkout payload and starts waiting', () => {
  const item = queueItem()

  assert.equal(item.status, 'menunggu')
  assert.equal(item.clientTxId, payload.clientTxId)
  assert.equal(item.storeId, 'store-17')
  assert.deepEqual(item.payload, payload)
  assert.deepEqual(item.summary, summary)
  assert.equal(item.createdAt, createdAt)
})

test('queue view contains only records from the active store', () => {
  const currentStoreItem = queueItem()
  const otherStoreItem = createOfflineQueueItem({ ...payload, clientTxId: '370e23b1-5d4c-4a64-a7b7-ecbcbfca94cb' }, summary, 'store-23', createdAt)

  assert.deepEqual(queueItemsForStore([currentStoreItem, otherStoreItem], storeId), [currentStoreItem])
})

test('retry keeps the original clientTxId and checkout payload', () => {
  const queued = queueItem()
  const failed = markQueueItemFailed(queued, 'Stok tidak cukup.', '2026-10-15T09:01:00.000Z')
  const retried = retryOfflineQueueItem(failed, '2026-10-15T09:02:00.000Z')

  assert.equal(retried.status, 'menunggu')
  assert.equal(retried.clientTxId, payload.clientTxId)
  assert.equal(retried.payload.clientTxId, payload.clientTxId)
  assert.deepEqual(retried.payload, payload)
  assert.equal(retried.error, undefined)
})

test('successful sync records the transaction status and server total', () => {
  const queued = queueItem()
  const synced = markQueueItemSynced(queued, {
    id: 'trx-41',
    status: 'completed',
    total: 18500,
  }, '2026-10-15T09:03:00.000Z')

  assert.equal(synced.status, 'tersinkron')
  assert.equal(synced.serverStatus, 'completed')
  assert.equal(synced.serverTransactionId, 'trx-41')
  assert.equal(synced.serverTotal, 18500)
  assert.equal(synced.syncedAt, '2026-10-15T09:03:00.000Z')
})

test('offline shortage is recorded as a synced server result, not a failed queue item', () => {
  const queued = queueItem()
  const synced = markQueueItemSynced(queued, {
    id: 'trx-42',
    status: 'completed_offline_shortage',
    total: 18500,
  }, '2026-10-15T09:03:00.000Z')

  assert.equal(synced.status, 'tersinkron')
  assert.equal(synced.serverStatus, 'completed_offline_shortage')
})

test('server errors move waiting items to actionable failure with the message', () => {
  const queued = queueItem()
  const failed = markQueueItemFailed(queued, 'Jumlah bayar tidak valid.', '2026-10-15T09:04:00.000Z')

  assert.equal(failed.status, 'gagal-perlu-tindakan')
  assert.equal(failed.error, 'Jumlah bayar tidak valid.')
  assert.equal(failed.failedAt, '2026-10-15T09:04:00.000Z')
})
