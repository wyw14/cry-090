import { describe, expect, it, vi } from 'vitest'
import { api } from './api'
describe('api adapter', () => { it('returns stable error message for failed request', async () => { vi.stubGlobal('fetch', vi.fn(async () => new Response(JSON.stringify({ message: '拒绝访问' }), { status: 403 }))); await expect(api.events()).rejects.toThrow('拒绝访问') }) })
