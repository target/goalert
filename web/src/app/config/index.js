import { isCypress } from '../env'

export const ITEMS_PER_PAGE = 15
export const POLL_INTERVAL = isCypress ? 1000 : 3500
export const POLL_ERROR_INTERVAL = isCypress ? 1000 : 30000

export const DEFAULT_SPIN_DELAY_MS = 200
export const DEFAULT_SPIN_WAIT_MS = 1500

export const DEBOUNCE_DELAY = 250

export const CREATE_ALERT_LIMIT = 35
export const BATCH_DELAY = 10
