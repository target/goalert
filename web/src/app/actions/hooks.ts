import { History, Location } from 'history'
import { useHistory, useLocation } from 'react-router-dom'

export type Value = string | boolean | number | string[]

// sanitizeURLParam serializes a value to be ready to store in a URL.
// If a value cannot or should not be stored, an empty string is returned.
export function sanitizeURLParam(value: Value): string | string[] {
  if (Array.isArray(value)) {
    const filtered = value.map((v) => v.trim()).filter(Boolean)
    if (filtered.length === 0) return ''
    return filtered
  }

  switch (typeof value) {
    case 'string':
      return value.trim()
    case 'boolean':
      if (value === true) return '1'
      return ''
    case 'number':
      return value.toString()
    default:
      return ''
  }
}

// getURLParam gets the URL param value and converts it from string|string[]
// to the desired type based on the provided default value.
export function getURLParam<T extends Value>(
  url: URLSearchParams,
  name: string,
  defaultValue: T,
): T {
  if (!url.has(name)) return defaultValue
  if (Array.isArray(defaultValue)) return url.getAll(name) as T
  if (typeof defaultValue === 'boolean') return (url.get(name) === '1') as T
  if (typeof defaultValue === 'string') return (url.get(name) || '') as T
  if (typeof defaultValue === 'number') return +(url.get(name) as string) as T
  return defaultValue
}

// setURL will replace the latest browser history entry with the provided config.
function setURL(
  config: URLSearchParams,
  location: Location,
  history: History,
): void {
  if (config.sort) config.sort()
  let newSearch = config.toString()
  newSearch = newSearch ? '?' + newSearch : ''

  if (newSearch === location.search) {
    // no action for no param change
    return
  }
  history.replace(location.pathname + newSearch + location.hash)
}

// useURLParam returns the value for a given URL param name if present, else the given default.
// It also returns a setter function that, when called, updates the URL param value.
// The native history stack will not push a new entry; instead, its
// latest entry will be replaced.
export function useURLParam<T extends Value>(
  name: string,
  defaultValue: T,
): [T, (newValue: T) => void] {
  const location = useLocation()
  const history = useHistory()
  const url = new URLSearchParams(location.search)

  const value = getURLParam<T>(url, name, defaultValue)

  function setValue(_newValue: T): void {
    const newValue =
      name === 'search' ? (_newValue as string) : sanitizeURLParam(_newValue)

    url.delete(name)

    if (Array.isArray(newValue)) {
      newValue.forEach((v) => url.append(name, v))
    } else if (newValue) {
      url.set(name, newValue)
    }

    setURL(url, location, history)
  }

  return [value, setValue]
}

// useResetURLParams returns a function that, when called, removes
// URL parameters for the given list of param names.
// The native history stack will not push a new entry; instead, its
// latest entry will be replaced.
export function useResetURLParams(...names: string[]): () => void {
  const location = useLocation()
  const history = useHistory()

  return function resetURLParams(): void {
    if (!names.length) {
      // nothing to do
      return
    }

    const url = new URLSearchParams(location.search)
    names.forEach((name) => url.delete(name))

    setURL(url, location, history)
  }
}
