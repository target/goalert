export const AUTH_LOGOUT = 'AUTH_LOGOUT'

// authLogout will update the user's auth state.
//
// If true is passed as an argument, a request to end
// the current session will be first made to the backend.
//
// AUTH_LOGOUT will be dispatched if, and after, the request completes.
export function authLogout(performFetch = false) {
  const payload = { type: AUTH_LOGOUT }
  if (!performFetch) return payload
  return dispatch =>
    fetch('/api/v2/identity/logout', {
      credentials: 'same-origin',
      method: 'POST',
    }).then(dispatch(payload))
}
