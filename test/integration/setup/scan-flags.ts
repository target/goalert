import { globSync } from 'glob'
import { readFileSync } from 'fs'
import { collectFlags } from '../lib'

// scanUniqueFlagCombos() is used to generate a list of unique flag combinations
// based on calls to baseURLFromFlags in the integration tests.
export function scanUniqueFlagCombos(): string[] {
  const flags = new Set<string>()
  const files = globSync('test/integration/**/*.spec.ts')

  for (const file of files) {
    const content = readFileSync(file, 'utf8')
    const flagSets = collectFlags(content)
    for (const flagSet of flagSets) flags.add(flagSet)
  }

  return Array.from(flags.keys()).sort()
}
