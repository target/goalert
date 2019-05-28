import { setURLParam, resetURLParams } from './main'

export const SET_ALERTS_CHECKED = 'SET_ALERTS_CHECKED'
export const SET_ALERTS_ACTION_COMPLETE = 'SET_ALERTS_ACTION_COMPLETE'

// setAlertsStatusFilter will set the current alert status filter.
// A falsy value will result in the default (active) being set.
export function setAlertsStatusFilter(type) {
  return setURLParam('filter', type && type !== 'active' ? type : null)
}

// setAlertsAllServicesFilter will set the alert list to include all services.
export function setAlertsAllServicesFilter(bool) {
  return setURLParam('allServices', bool ? '1' : null)
}

// resetAlertsFilters will reset all alert list filters to their defaults (NOT including search).
export function resetAlertsFilters() {
  return resetURLParams('filter', 'allServices')
}

export function setCheckedAlerts(array) {
  return {
    type: SET_ALERTS_CHECKED,
    payload: array,
  }
}

export function setAlertsActionComplete(bool) {
  return {
    type: SET_ALERTS_ACTION_COMPLETE,
    payload: bool,
  }
}
