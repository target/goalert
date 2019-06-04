import { replace } from 'connected-react-router'
export const SET_SHOW_NEW_USER_FORM = 'SET_SHOW_NEW_USER_FORM'

// resetURLParams will reset all url parameters.
//
// An optional list of specific keys to reset can be passed.
export function resetURLParams(...keys) {
  return (dispatch, getState) => {
    const state = getState()
    if (!keys.length) return dispatch(replace(state.router.location.pathname))

    const q = new URLSearchParams(state.router.location.search)
    keys.forEach(key => {
      q.delete(key)
    })
    if (q.sort) q.sort()

    const search = q.toString()
    dispatch(
      replace(state.router.location.pathname + (search ? '?' + search : '')),
    )
  }
}

const sanitizeParam = value => {
  if (value === true) value = '1' // explicitly true
  if (!value) value = '' // any falsey value
  if (!Array.isArray(value)) return value.trim()

  let filtered = value.filter(v => v)
  if (filtered.length === 0) return null

  return filtered
}

// setSearch will set the current search parameter/filter.
export function setSearch(value) {
  return setURLParam('search', value || '')
}

// setURLParam will update the URL parameter with the given name to the provided value.
//
// Falsy values will result in the parameter being cleared.
// The value can also be an array of strings. An empty array will result in the parameter
// being cleared.
export function setURLParam(name, _value, _default) {
  return (dispatch, getState) => {
    const state = getState()

    if (_value === _default) {
      _value = ''
    }
    const value = sanitizeParam(_value)

    const q = new URLSearchParams(state.router.location.search)
    if (Array.isArray(value)) {
      q.delete(name)
      value.forEach(v => q.append(name, v))
    } else if (value) {
      q.set(name, value)
    } else {
      q.delete(name)
    }
    if (q.sort) q.sort()

    const search = q.toString()
    dispatch(
      replace(state.router.location.pathname + (search ? '?' + search : '')),
    )
  }
}

export function setShowNewUserForm(search) {
  return {
    type: SET_SHOW_NEW_USER_FORM,
    payload: search,
  }
}
