import { useHistory, useLocation } from 'react-router-dom'
import { urlParamSelector, urlPathSelector } from '../selectors'
import { setURLParam as setParam } from './main'
import { useSelector, useDispatch } from 'react-redux'
import { warn } from '../util/debug'
import joinURL from '../util/joinURL'
import { pathPrefix } from '../env'

export type Value = string | boolean | number | string[]

export function useURLParam<T extends Value>(
  name: string,
  defaultValue: T,
): [T, (newValue: T) => void] {
  const dispatch = useDispatch()
  const urlParam = useSelector(urlParamSelector)
  const urlPath = joinURL(pathPrefix, useSelector(urlPathSelector))
  const value = urlParam(name, defaultValue) as T

  function setValue(newValue: T): void {
    if (window.location.pathname !== urlPath) {
      warn(
        'useURLParam was called to set a parameter, but location.pathname has changed, aborting',
      )
      return
    }
    dispatch(setParam(name, newValue, defaultValue))
  }

  return [value, setValue]
}

// useResetURLParams returns a function that, when called, will remove
// the query parameters for a given list of key names. If no list is
// provided, all existing query paramters are removed.
// The native history stack will not push a new entry; instead, its
// latest entry will be replaced.
export function useResetURLParams(...keys: string[]): () => void {
  const { pathname, search, hash } = useLocation()
  const history = useHistory()

  return function resetURLParams(): void {
    if (!keys.length) {
      // by default, clear all params
      return history.replace(pathname)
    }

    const q = new URLSearchParams(search)
    keys.forEach((key) => q.delete(key))

    if (q.sort) {
      q.sort()
    }

    let newSearch = q.toString()
    if (newSearch) {
      newSearch = '?' + newSearch
    }

    if (newSearch === search) {
      // no action for no param change
      return
    }

    history.replace(pathname + newSearch + hash)
  }
}
