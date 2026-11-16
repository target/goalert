// withPathPrefix prepends the public URL path prefix to an absolute in-app
// path. External URLs and relative paths are returned unchanged.
export default function withPathPrefix(to: string, prefix: string): string {
  if (!to.startsWith('/') || to.startsWith('//')) return to

  return prefix.replace(/\/$/, '') + to
}
