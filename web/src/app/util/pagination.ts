import { useState } from 'react'

/**
 * usePages is a custom hook that manages pagination state by tracking the current page cursor
 * as well as previous page cursors.
 *
 * @returns {(() => string) | undefined}  A function to go back to the previous page, or undefined if there is no previous page.
 * @returns {(() => string) | undefined} A function to go to the next page, or undefined if there is no next page.
 * @returns {() => string} A function to reset the page cursor to the first page.
 */
export function usePages(
  nextCursor: string | null | undefined,
): [(() => string) | undefined, (() => string) | undefined, () => string] {
  const [pageCursors, setPageCursors] = useState([''])

  function goBack(): string {
    const newCursors = pageCursors.slice(0, -1)
    setPageCursors(newCursors)
    return newCursors[newCursors.length - 1]
  }

  function goNext(): string {
    if (!nextCursor) return pageCursors[pageCursors.length - 1]
    setPageCursors([...pageCursors, nextCursor])
    return nextCursor
  }

  return [
    pageCursors.length > 1 ? goBack : undefined,
    nextCursor ? goNext : undefined,
    () => {
      setPageCursors([''])
      return ''
    },
  ]
}
