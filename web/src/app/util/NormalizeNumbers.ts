export function normalizeNumbers(
  val: number,
  min: number,
  max: number,
): number {
  if (val < min) return min
  if (val > max) return max
  return val
}
