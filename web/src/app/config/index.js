export const ITEMS_PER_PAGE = 15
export default {
  BASE_API_URL: process.env.GO_ALERT_BASE_API_URL || '/api',
}
export const POLL_INTERVAL = global.Cypress ? 1000 : 3500
export const POLL_ERROR_INTERVAL = global.Cypress ? 1000 : 30000

export const DEFAULT_SPIN_DELAY_MS = 200
export const DEFAULT_SPIN_WAIT_MS = 1500

export const DEBOUNCE_DELAY = global.Cypress ? 50 : 250
