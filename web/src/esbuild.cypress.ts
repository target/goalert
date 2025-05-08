#!/usr/bin/env node
/* eslint @typescript-eslint/no-var-requires: 0 */
/* eslint @typescript-eslint/no-require-imports: 0 */
const glob = require('glob')

const intEntry = {}
glob.globSync(path.join(__dirname, 'cypress/e2e/*')).forEach((file) => {
  const name = path.basename(file, '.ts')
  intEntry['integration/' + name] = file
})

async function run(): Promise<void> {
  const method = process.argv.includes('--watch') ? 'context' : 'build'

  const ctx = await require('esbuild')[method]({
    entryPoints: {
      'support/index': path.join(__dirname, 'cypress/support/e2e.ts'),
      ...intEntry,
    },
    outdir: 'bin/build/integration/cypress',
    logLevel: 'info',
    bundle: true,
    define: { global: 'window' },
    minify: true,
    sourcemap: 'linked',
    target: ['chrome80', 'firefox99', 'safari12', 'edge79'],
  })

  if (process.argv.includes('--watch')) {
    await ctx.watch()
  }
}

run().catch((err) => {
  console.error(err)
  process.exit(1)
})
