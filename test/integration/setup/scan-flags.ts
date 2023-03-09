import { globSync } from 'glob'
import { readFileSync } from 'fs'

const callRx = /baseURLFromFlags\(([^)]+)\)/g
const validFlagsRx = /^[a-z,-]+$/

// scanUniqueFlagCombos() is used to generate a list of unique flag combinations
// based on calls to baseURLFromFlags in the integration tests.
export function scanUniqueFlagCombos(): string[] {
  const flags = new Set<string>()
  const files = globSync('test/integration/**/*.spec.ts')

  for (const file of files) {
    const content = readFileSync(file, 'utf8')
    const m = callRx.exec(content)
    if (!m) continue

    m.slice(1).forEach((match) => {
      // The regex will include the entire parameter, including the surrounding
      // square brackets, so we can just parse it as JSON after fixing the
      // quotes.
      const items = JSON.parse(match.replace(/'/g, '"')) as string[]
      if (!Array.isArray(items)) throw new Error('not array in ' + file)

      const set = items.sort().join(',')
      if (!validFlagsRx.test(set)) throw new Error('invalid flags in ' + file)

      flags.add(set)
    })
  }

  return Array.from(flags.keys()).sort()
}
