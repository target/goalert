// calcNewActiveIndex returns the newActiveIndex for a swap operation
// -1 will be returned if there was no change
export function calcNewActiveIndex(
  oldActiveIndex: number,
  oldIndex: number,
  newIndex: number,
): number {
  if (oldIndex === newIndex) {
    return -1
  }
  if (oldActiveIndex === oldIndex) {
    return newIndex
  }

  if (oldIndex > oldActiveIndex && newIndex <= oldActiveIndex) {
    return oldActiveIndex + 1
  }

  if (oldIndex < oldActiveIndex && newIndex >= oldActiveIndex) {
    return oldActiveIndex - 1
  }
  return -1
}

// reorderList will move an item from the oldIndex to the newIndex, preserving order
// returning the result as a new array.
export function reorderList<T>(
  _items: T[],
  oldIndex: number,
  newIndex: number,
): T[] {
  const items = _items.slice()
  items.splice(oldIndex, 1) // remove 1 element from oldIndex position
  items.splice(newIndex, 0, _items[oldIndex]) // add dest to newIndex position
  return items
}
