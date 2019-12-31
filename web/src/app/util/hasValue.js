// cast as a Boolean unless candidate is an Array
export default function hasValue(candidate) {
  if (Array.isArray(candidate)) {
    return candidate.length > 0
  }
  return Boolean(candidate)
}
