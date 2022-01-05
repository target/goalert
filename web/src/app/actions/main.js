export const SET_SHOW_NEW_USER_FORM = 'SET_SHOW_NEW_USER_FORM'

export function setShowNewUserForm(search) {
  return {
    type: SET_SHOW_NEW_USER_FORM,
    payload: search,
  }
}
