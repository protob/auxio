import { request } from './client'
import { Stats } from './schemas'

export function getStats() {
  return request('/stats', Stats)
}
