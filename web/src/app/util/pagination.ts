import { useState } from 'react'

/**
 * usePages is a custom hook that manages pagination state by tracking the current page cursor
 * as well as previous page cursors.
 *
 * @returns {string} The current page cursor.
 * @returns {() => void | undefined}  A function to go back to the previous page, or undefined if there is no previous page.
 * @returns {(nextCursor: string | null | undefined) => (() => void) | undefined} A function to generate a callback to go to the next page, or undefined if there is no next page.
 */
export function usePages(): [
  string,
  (() => void) | undefined,
  (nextCursor: string | null | undefined) => (() => void) | undefined,
] {
  const [pageCursors, setPageCursors] = useState([''])

  return [
    pageCursors[pageCursors.length - 1],
    pageCursors.length > 1
      ? () => setPageCursors(pageCursors.slice(0, -1))
      : undefined,
    (nextCursor) =>
      nextCursor
        ? () => setPageCursors([...pageCursors, nextCursor])
        : undefined,
  ]
}
