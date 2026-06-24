import { useLocation, useSearch } from 'wouter'

export type Value = string | boolean | number | string[]

export function useURLKey(): string {
  const [path] = useLocation()
  const search = useSearch()
  return path + search
}

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

// getParamValues converts each URL search param into the
// desired type based on its respective default value.
export function getParamValues<T extends Record<string, Value>>(
  search: string,
  params: T, // <name, default> pairs
): T {
  const result = {} as Record<string, Value>
  const q = new URLSearchParams(search)

  for (const [name, defaultVal] of Object.entries(params)) {
    if (!q.has(name)) {
      result[name] = defaultVal
    } else if (Array.isArray(defaultVal)) {
      result[name] = q.getAll(name)
    } else if (typeof defaultVal === 'boolean') {
      result[name] = q.get(name) === '1'
    } else if (typeof defaultVal === 'string') {
      result[name] = q.get(name) || ''
    } else if (typeof defaultVal === 'number') {
      result[name] = +(q.get(name) as string)
    } else {
      result[name] = defaultVal
    }
  }
  return result as T
}

// newSearch will return a normalized URL search string if different from the current one.
function newSearch(params: URLSearchParams): [boolean, string] {
  if (params.sort) params.sort()
  let newSearch = params.toString()
  newSearch = newSearch ? '?' + newSearch : ''
  if (newSearch === location.search) {
    // no action for no param change
    return [false, '']
  }

  return [true, newSearch]
}

// useURLParams returns the values for the given URL params if present, else the given defaults.
// It also returns a setter function that, when called, updates the URL param values.
// The native history stack will not push a new entry; instead, its
// latest entry will be replaced.
export function useURLParams<T extends Record<string, Value>>(
  params: T, // <name, default> pairs
): [T, (newValues: Partial<T>) => void] {
  const [path, navigate] = useLocation()
  const search = useSearch()
  const q = new URLSearchParams(search)
  let called = false

  function setParams(newParams: Partial<T>): void {
    if (called) {
      console.error(
        'useURLParams: setParams was called multiple times in one render, aborting',
      )
      return
    }
    called = true

    for (const [name, _v] of Object.entries(newParams)) {
      const value = name === 'search' ? (_v as string) : sanitizeURLParam(_v)

      q.delete(name)

      if (Array.isArray(value)) {
        value.forEach((v) => q.append(name, v))
      } else if (value) {
        q.set(name, value)
      }
    }

    const [hasNew, search] = newSearch(q)
    if (!hasNew) {
      // nothing to do
      return
    }

    navigate(path + search + location.hash, { replace: true })
  }

  const values = getParamValues<T>(search, params)

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
  const [path, navigate] = useLocation()
  const searchVal = useSearch()
  let called = false

  return function resetURLParams(): void {
    if (!names.length) {
      // nothing to do
      return
    }

    if (called) {
      console.error(
        'useResetURLParams: resetURLParams was called multiple times in one render, aborting',
      )
      return
    }
    called = true

    const params = new URLSearchParams(searchVal)
    names.forEach((name) => params.delete(name))

    const [hasNew, search] = newSearch(params)
    if (!hasNew) {
      // nothing to do
      return
    }

    navigate(path + search + location.hash, { replace: true })
  }
}
