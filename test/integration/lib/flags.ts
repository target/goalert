import cfg from '../../../playwright.config.js'

export function baseURLFromFlags(flags: string[]): string {
  const flagStr = flags.sort().join(',')

  const srv = cfg.webServer.find((ws) =>
    ws.command.includes('--experimental=' + flagStr),
  )
  if (!srv || !srv.url) {
    throw new Error(
      `No valid web server configured with experimental flags: ${flags.join(
        ',',
      )}`,
    )
  }

  return srv.url.replace(/\/health$/, '')
}

const validFlagRx = /^[a-z-]+$/

// collectFlags will return a list of all unique flag combinations used in the
// provided data.
export function collectFlags(data: string): Array<string> {
  const flags = new Set<string>()

  while (true) {
    const idx = data.indexOf('baseURLFromFlags(')
    if (idx < 0) break

    data = data.slice(idx + 17)
    const end = data.indexOf(')')
    if (end < 0) break

    const flagSet = JSON.parse(data.slice(0, end).replace(/'/g, '"'))
    if (!Array.isArray(flagSet)) continue

    flagSet.sort().forEach((f) => {
      if (!validFlagRx.test(f)) throw new Error('invalid flag ' + f)
    })
    flags.add(flagSet.sort().join(','))

    data = data.slice(end)
  }

  return Array.from(flags.keys()).sort()
}
