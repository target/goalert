import Fuse from 'fuse.js'
import { useEffect, useRef } from 'react'

interface FuseParams<T> {
  data: T[]
  keys?: Fuse.FuseOptionKey<T>[]
  search: string
  options?: Fuse.IFuseOptions<T> & CustomOptions
}

interface CustomOptions {
  showResultsWhenNoSearchTerm?: boolean
}

const defaultOptions = {
  shouldSort: true,
  threshold: 0.1,
  location: 0,
  distance: 100,
  maxPatternLength: 32,
  minMatchCharLength: 1,
}

const DEFAULT_QUERY = ''

export function useFuse<T>({
  data,
  keys,
  search = DEFAULT_QUERY,
  options = {},
}: FuseParams<T>): Fuse.FuseResult<T>[] {
  const { showResultsWhenNoSearchTerm, ...fuseOptions } = options
  const fuse = useRef<Fuse<T>>()

  useEffect(() => {
    if (!data) return

    fuse.current = new Fuse(data, {
      ...defaultOptions,
      ...fuseOptions,
      keys,
    })
  }, [fuse, data, keys, fuseOptions])

  const fuseResults = fuse.current ? fuse.current.search(search) : []

  const results =
    showResultsWhenNoSearchTerm && search === ''
      ? data.map((data, i) => ({
          item: data,
          score: 1,
          refIndex: i,
        }))
      : fuseResults

  return results
}
