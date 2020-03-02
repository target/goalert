import { setURLParam, resetURLParams } from './main'

export const SET_ALERTS_ACTION_COMPLETE = 'SET_ALERTS_ACTION_COMPLETE'

// setAlertsStatusFilter will set the current alert status filter.
// A falsy value will result in the default (active) being set.
// active: null
// unacknowledged
// acknowledged
// closed
// all
export function setAlertsStatusFilter(type) {
  let val = null
  if (type && type !== 'active') {
    // active is the default, so when type is null it will show the "active" tab
    val = type
  }

  return setURLParam('filter', val)
}

// setAlertsAllServicesFilter will set the alert list to include all services.
export function setAlertsAllServicesFilter(bool) {
  return setURLParam('allServices', bool ? '1' : null)
}

// resetAlertsFilters will reset all alert list filters to their defaults (NOT including search).
export function resetAlertsFilters() {
  return resetURLParams('filter', 'allServices')
}

export function setAlertsActionComplete(bool) {
  return {
    type: SET_ALERTS_ACTION_COMPLETE,
    payload: bool,
  }
}
