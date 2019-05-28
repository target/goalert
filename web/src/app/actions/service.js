export const SET_SERVICE_SEARCH = 'SET_SERVICE_SEARCH'

export function setServiceSearch(search) {
  return {
    type: SET_SERVICE_SEARCH,
    payload: search,
  }
}
