#!/usr/bin/env node
/* eslint @typescript-eslint/no-var-requires: 0 */
const path = require('path')
const glob = require('glob')

const intEntry = {}
glob.globSync(path.join(__dirname, 'cypress/e2e/*')).forEach((file) => {
  const name = path.basename(file, '.ts')
  intEntry['integration/' + name] = file
})

async function run() {
  const ctx = await require('esbuild').context({
    entryPoints: {
      'support/index': 'web/src/cypress/support/e2e.ts',
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
  } else {
    await ctx.dispose()
  }
}

run().catch((err) => {
  console.error(err)
  process.exit(1)
})
