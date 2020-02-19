export default function joinURL(...parts) {
  parts = parts.filter(p => p) // remove empty segments
  if (!parts || parts.length === 0) return ''

  return parts
    .join('/')
    .replace(/\/\/+/g, '/')
    .replace(/\/$/, '')
    .split('/')
    .filter((part, idx, parts) => {
      if (idx === 0) return true
      if (part === '.' || part === '..') return false
      if (parts[idx + 1] === '..') return false
      return true
    })
    .join('/')
}
