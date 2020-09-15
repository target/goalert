import { isCypress } from '../env'
import { Duration } from 'luxon'

export const ITEMS_PER_PAGE = 15
export const POLL_INTERVAL = isCypress ? 1000 : 3500
export const POLL_ERROR_INTERVAL = isCypress ? 1000 : 30000

export const DEFAULT_SPIN_DELAY_MS = 200
export const DEFAULT_SPIN_WAIT_MS = 1500

export const DEBOUNCE_DELAY = 250

export const CREATE_ALERT_LIMIT = 35
export const BATCH_DELAY = 10

// UPDATE_CHECK_INTERVAL controls how often we poll for updates.
export const UPDATE_CHECK_INTERVAL = Duration.fromObject({ minutes: 1 })

// UPDATE_NOTIF_DURATION controls how long a new version must be seen (and stable) before
// displaying the persistent update notification.
export const UPDATE_NOTIF_DURATION = Duration.fromObject({ minutes: 15 })

// UPDATE_FORCE_DURATION controls how long a new version must been seen (and stable) before
// the page is forcibly refreshed.
export const UPDATE_FORCE_DURATION = Duration.fromObject({ hours: 3 })
