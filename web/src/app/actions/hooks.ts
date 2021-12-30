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

// getParamValues converts each value from string|string[]
// into the desired type based on the respective default value.
export function getParamValues<T extends Record<string, Value>>(
  location: Location,
  params: T,
): T {
  const result = {} as Record<string, Value>
  const q = new URLSearchParams(location.search)

  for (const [k, defaultv] of Object.entries(params)) {
    if (!q.has(k)) {
      result[k] = defaultv
    } else if (Array.isArray(defaultv)) {
      result[k] = q.getAll(k)
    } else if (typeof defaultv === 'boolean') {
      result[k] = q.get(k) === '1'
    } else if (typeof defaultv === 'string') {
      result[k] = q.get(k) || ''
    } else if (typeof defaultv === 'number') {
      result[k] = +(q.get(k) as string)
    } else {
      result[k] = defaultv
    }
  }
  return result as T
}

// setURLParams will replace the latest browser history entry with the provided params.
function setURLParams(
  history: History,
  location: Location,
  params: URLSearchParams,
): void {
  if (params.sort) params.sort()
  let newSearch = params.toString()
  newSearch = newSearch ? '?' + newSearch : ''

  if (newSearch === location.search) {
    // no action for no param change
    return
  }
  history.replace(location.pathname + newSearch + location.hash)
}

// useURLParams returns the values for the given URL params if present, else the given defaults.
// It also returns a setter function that, when called, updates the URL param values.
// The native history stack will not push a new entry; instead, its
// latest entry will be replaced.
export function useURLParams<T extends Record<string, Value>>(
  params: T, // <name, default> pairs
): [T, (newValues: Partial<T>) => void] {
  const location = useLocation()
  const history = useHistory()
  const q = new URLSearchParams(location.search)
  let called = false

  function setParams(newParams: Partial<T>): void {
    if (called) {
      console.error(
        'useURLParams: setParams was called multiple times in one render, aborting',
      )
      return
    }
    called = true

    for (const [k, _v] of Object.entries(newParams)) {
      const v = k === 'search' ? (_v as string) : sanitizeURLParam(_v)

      q.delete(k)

      if (Array.isArray(v)) {
        v.forEach((v) => q.append(k, v))
      } else if (v) {
        q.set(k, v)
      }
    }

    setURLParams(history, location, q)
  }

  const values = getParamValues<T>(location, params)

  return [values, setParams]
}

// useURLParam is like useURLParams but only handles one parameter.
export function useURLParam<T extends Value>(
  name: string,
  defaultValue: T,
): [T, (newValue: T) => void] {
  const [params, setParams] = useURLParams({ [name]: defaultValue })

  function setValue(newValue: T): void {
    setParams({ [name]: newValue })
  }

  return [params[name], setValue]
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

    const params = new URLSearchParams(location.search)
    names.forEach((name) => params.delete(name))

    setURLParams(history, location, params)
  }
}
