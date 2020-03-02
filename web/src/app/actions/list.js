export const SET_CHECKED_LIST_ITEMS = 'SET_CHECKED_LIST_ITEMS'

export function setCheckedItems(array) {
  return {
    type: SET_CHECKED_LIST_ITEMS,
    payload: array,
  }
}
