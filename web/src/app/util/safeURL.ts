// encodeHtmlEntities will encode common HTML entities in a string.
//
// This is useful for ensuring that comparisons between html encoded urls and
// non-encoded labels are accurate
function encodeHtmlEntities(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

// safeURL will determine if a url is safe for linking.
//
// It tries to determine if the label is misleading.
export function safeURL(url: string, label: string): boolean {
  label = encodeHtmlEntities(label)

  if (url.startsWith('mailto:')) {
    const email = url.substr(7)
    return email === label && email.includes('@')
  }

  if (url.startsWith('tel:')) {
    const phone = url.substr(4)
    return phone === label && /^\+?[\d\- ]+$/.test(phone)
  }

  // handle http protocols
  if (!/https?:\/\//.test(url)) return false // require absolute URLs
  if (!/[./]/.test(label)) return true // don't consider it a path/url without slashes or periods
  if (url.startsWith(label)) return true // if it matches the beginning, then it's fine
  if (url.replace(/^https?:\/\//, '').startsWith(label)) return true // same prefix without protocol
  if (url.replace(/^https?:\/\//, '').startsWith('www.' + label)) return true // same prefix without protocol

  return false
}
