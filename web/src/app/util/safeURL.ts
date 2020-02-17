// safeURL will determine if a url is safe for linking.
//
// It tries to determine if the label is misleading.
export function safeURL(url: string, label: string) {
  if (!/https?:\/\//.test(url)) return false // require absolute URLs
  if (!/[./]/.test(label)) return true // don't consider it a path/url without slashes or periods
  if (url.startsWith(label)) return true // if it matches the begining, then it's fine
  if (url.replace(/^https?:\/\//, '').startsWith(label)) return true // same prefix without protocol
  if (url.replace(/^https?:\/\//, '').startsWith('www.' + label)) return true // same prefix without protocol

  return false
}
