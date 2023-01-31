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
