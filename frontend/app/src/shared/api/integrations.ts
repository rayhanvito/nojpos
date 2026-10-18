// Endpoint integrasi: kredensial API dan langganan webhook admin.

import { request } from './core.ts'
import type { Page } from './types.ts'

export interface ApiKey {
  id: string | number
  name: string
  prefix: string
  created_at: string
  last_used_at: string | null
}

export type ApiKeyCreated = ApiKey & { apiKey: string }

export type WebhookEvent = 'transaction.created' | 'transaction.refunded' | 'stock.updated'

export interface Webhook {
  id: string | number
  url: string
  events: WebhookEvent[]
  created_at: string
}

export type WebhookInput = { url: string; events: WebhookEvent[] }
export type WebhookCreated = Webhook & { secret: string }

export function apiListApiKeys() {
  return request<Page<ApiKey>>('GET', '/api-keys')
}

export function apiCreateApiKey(name: string) {
  return request<ApiKeyCreated>('POST', '/api-keys', { name })
}

export function apiDeleteApiKey(id: string | number) {
  return request<{ message: string }>('DELETE', `/api-keys/${id}`)
}

export function apiListWebhooks() {
  return request<Page<Webhook>>('GET', '/webhooks')
}

export function apiCreateWebhook(body: WebhookInput) {
  return request<WebhookCreated>('POST', '/webhooks', body)
}

export function apiDeleteWebhook(id: string | number) {
  return request<{ message: string }>('DELETE', `/webhooks/${id}`)
}
