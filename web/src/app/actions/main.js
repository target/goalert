import { replace } from 'connected-react-router'
export const SET_SHOW_NEW_USER_FORM = 'SET_SHOW_NEW_USER_FORM'

const setSearchStr = (dispatch, state, searchParams) => {
  if (searchParams.sort) searchParams.sort()
  let newSearch = searchParams.toString()
  newSearch = newSearch ? '?' + newSearch : ''

  const { search, pathname, hash } = state.router.location
  if (newSearch === search) {
    // no action for no param change
    return
  }
  dispatch(replace(pathname + newSearch + hash))
}

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

    setSearchStr(dispatch, state, q)
  }
}
const sanitizeParam = value => {
  if (value === true) value = '1' // explicitly true
  if (!value) value = '' // any falsey value
  if (!Array.isArray(value)) return value.trim()

  const filtered = value.filter(v => v)
  if (filtered.length === 0) return null

  return filtered
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
    const value = name === 'search' ? _value : sanitizeParam(_value)

    const q = new URLSearchParams(state.router.location.search)
    if (Array.isArray(value)) {
      q.delete(name)
      value.forEach(v => q.append(name, v))
    } else if (value) {
      q.set(name, value)
    } else {
      q.delete(name)
    }

    setSearchStr(dispatch, state, q)
  }
}

// setSearch will set the current search parameter/filter.
export function setSearch(value) {
  return setURLParam('search', value || '')
}

export function setShowNewUserForm(search) {
  return {
    type: SET_SHOW_NEW_USER_FORM,
    payload: search,
  }
}
