import type { CheckoutBody, Trx } from './api'

type OfflineQueueStatus = 'menunggu' | 'tersinkron' | 'gagal-perlu-tindakan'
type OfflineQueuePayload = CheckoutBody & { clientTxId: string }

export interface OfflineQueueSummary {
  items: { name: string; qty: number }[]
  estimatedTotal: number
}

export interface OfflineQueueItem {
  storeId: string
  clientTxId: string
  payload: OfflineQueuePayload
  summary: OfflineQueueSummary
  status: OfflineQueueStatus
  createdAt: string
  serverStatus?: Trx['status']
  serverTransactionId?: string
  serverTotal?: number
  syncedAt?: string
  lastAttemptAt?: string
  error?: string
  failedAt?: string
}

type CheckoutSender = (payload: CheckoutBody) => Promise<Trx>

const DB_NAME = 'nojpos-main-offline-queue'
const DB_VERSION = 1
const STORE_NAME = 'transactions'
const CHANGE_EVENT = 'op:offline-queue-change'

let database: Promise<IDBDatabase> | null = null
let syncing: Promise<void> | null = null

export function createClientTxId(): string {
  const cryptoApi = globalThis.crypto
  if (typeof cryptoApi?.randomUUID === 'function') return cryptoApi.randomUUID()
  const bytes = new Uint8Array(16)
  if (cryptoApi?.getRandomValues) cryptoApi.getRandomValues(bytes)
  else for (let i = 0; i < bytes.length; i++) bytes[i] = Math.floor(Math.random() * 256)
  bytes[6] = (bytes[6]! & 0x0f) | 0x40
  bytes[8] = (bytes[8]! & 0x3f) | 0x80
  const hex = [...bytes].map((byte) => byte.toString(16).padStart(2, '0')).join('')
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`
}

export function isNetworkFailure(error: unknown, isOnline: boolean): boolean {
  if (error !== null && typeof error === 'object' && 'status' in error) return false
  return !isOnline || error instanceof TypeError
}

export function createOfflineQueueItem(
  payload: CheckoutBody,
  summary: OfflineQueueSummary,
  storeId: string,
  createdAt = new Date().toISOString(),
): OfflineQueueItem {
  const clientTxId = payload.clientTxId || createClientTxId()
  return {
    storeId,
    clientTxId,
    payload: { ...payload, clientTxId },
    summary,
    status: 'menunggu',
    createdAt,
  }
}

export function queueItemsForStore(items: OfflineQueueItem[], storeId: string): OfflineQueueItem[] {
  return items.filter((item) => item.storeId === storeId)
}

export function retryOfflineQueueItem(item: OfflineQueueItem, retriedAt = new Date().toISOString()): OfflineQueueItem {
  if (item.status !== 'gagal-perlu-tindakan') return item
  return { ...item, status: 'menunggu', lastAttemptAt: retriedAt, error: undefined, failedAt: undefined }
}

export function markQueueItemSynced(item: OfflineQueueItem, trx: Pick<Trx, 'id' | 'status' | 'total'>, syncedAt = new Date().toISOString()): OfflineQueueItem {
  return {
    ...item,
    status: 'tersinkron',
    serverStatus: trx.status,
    serverTransactionId: trx.id,
    serverTotal: trx.total,
    syncedAt,
    error: undefined,
    failedAt: undefined,
  }
}

export function markQueueItemFailed(item: OfflineQueueItem, message: string, failedAt = new Date().toISOString()): OfflineQueueItem {
  return { ...item, status: 'gagal-perlu-tindakan', error: message, failedAt }
}

function openDatabase(): Promise<IDBDatabase> {
  if (database) return database
  if (typeof indexedDB === 'undefined') return Promise.reject(new Error('Penyimpanan antrean tidak tersedia di browser ini.'))
  const opening = new Promise<IDBDatabase>((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION)
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains(STORE_NAME)) db.createObjectStore(STORE_NAME, { keyPath: ['storeId', 'clientTxId'] })
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error ?? new Error('Gagal membuka penyimpanan antrean.'))
    request.onblocked = () => reject(new Error('Penyimpanan antrean sedang dipakai tab lain. Muat ulang halaman lalu coba lagi.'))
  }).catch((error: unknown) => {
    database = null
    throw error
  })
  database = opening
  return opening
}

function transactionDone(transaction: IDBTransaction): Promise<void> {
  return new Promise((resolve, reject) => {
    transaction.oncomplete = () => resolve()
    transaction.onerror = () => reject(transaction.error ?? new Error('Gagal menyimpan antrean transaksi.'))
    transaction.onabort = () => reject(transaction.error ?? new Error('Penyimpanan antrean transaksi dibatalkan.'))
  })
}

function requestValue<T>(request: IDBRequest<T>): Promise<T> {
  return new Promise((resolve, reject) => {
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error ?? new Error('Gagal membaca antrean transaksi.'))
  })
}

function notifyQueueChanged() {
  if (typeof window !== 'undefined') window.dispatchEvent(new Event(CHANGE_EVENT))
}

async function saveOfflineQueueItem(item: OfflineQueueItem): Promise<void> {
  const db = await openDatabase()
  const tx = db.transaction(STORE_NAME, 'readwrite')
  const done = transactionDone(tx)
  tx.objectStore(STORE_NAME).put(item)
  await done
  notifyQueueChanged()
}

export async function enqueueOfflineCheckout(payload: CheckoutBody, summary: OfflineQueueSummary, storeId: string | undefined): Promise<OfflineQueueItem> {
  if (!storeId) throw new Error('Toko aktif tidak tersedia untuk menyimpan antrean.')
  const item = createOfflineQueueItem(payload, summary, storeId)
  await saveOfflineQueueItem(item)
  return item
}

export async function listOfflineQueue(): Promise<OfflineQueueItem[]> {
  const db = await openDatabase()
  const tx = db.transaction(STORE_NAME, 'readonly')
  const done = transactionDone(tx)
  const items = await requestValue(tx.objectStore(STORE_NAME).getAll() as IDBRequest<OfflineQueueItem[]>)
  await done
  return items.sort((a, b) => a.createdAt.localeCompare(b.createdAt))
}

export async function dismissOfflineQueueItem(storeId: string, clientTxId: string): Promise<boolean> {
  const db = await openDatabase()
  const tx = db.transaction(STORE_NAME, 'readwrite')
  const done = transactionDone(tx)
  const store = tx.objectStore(STORE_NAME)
  const item = await requestValue(store.get([storeId, clientTxId]) as IDBRequest<OfflineQueueItem | undefined>)
  if (!item || item.status === 'menunggu') {
    tx.abort()
    await done.catch(() => {})
    return false
  }
  store.delete([storeId, clientTxId])
  await done
  notifyQueueChanged()
  return true
}

export async function retryOfflineQueueItemInStore(storeId: string, clientTxId: string): Promise<OfflineQueueItem | null> {
  const db = await openDatabase()
  const tx = db.transaction(STORE_NAME, 'readwrite')
  const done = transactionDone(tx)
  const store = tx.objectStore(STORE_NAME)
  const item = await requestValue(store.get([storeId, clientTxId]) as IDBRequest<OfflineQueueItem | undefined>)
  if (!item || item.status !== 'gagal-perlu-tindakan') {
    tx.abort()
    await done.catch(() => {})
    return null
  }
  const retried = retryOfflineQueueItem(item)
  store.put(retried)
  await done
  notifyQueueChanged()
  return retried
}

function errorMessage(error: unknown): string {
  return error instanceof Error && error.message ? error.message : 'Gagal menyinkronkan transaksi.'
}

export function syncOfflineQueue(checkout: CheckoutSender, storeId: string | undefined): Promise<void> {
  if (syncing) return syncing
  if (!storeId || (typeof navigator !== 'undefined' && !navigator.onLine)) return Promise.resolve()
  syncing = syncQueue(checkout, storeId).finally(() => { syncing = null })
  return syncing
}

async function syncQueue(checkout: CheckoutSender, storeId: string): Promise<void> {
  const items = queueItemsForStore(await listOfflineQueue(), storeId).filter((item) => item.status === 'menunggu')
  if (!items.length) return
  for (const item of items) {
    if (typeof navigator !== 'undefined' && !navigator.onLine) return
    let trx: Trx
    try {
      trx = await checkout(item.payload)
    } catch (error) {
      if (isNetworkFailure(error, typeof navigator === 'undefined' || navigator.onLine)) return
      await saveOfflineQueueItem(markQueueItemFailed(item, errorMessage(error)))
      continue
    }
    await saveOfflineQueueItem(markQueueItemSynced(item, trx))
  }
}
