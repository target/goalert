#!/usr/bin/env node
/* eslint @typescript-eslint/no-var-requires: 0 */
const path = require('path')
const glob = require('glob')

const intEntry = {}
glob.sync(path.join(__dirname, 'cypress/integration/*')).forEach((file) => {
  const name = path.basename(file, '.ts')
  intEntry['integration/' + name] = file
})

require('esbuild')
  .build({
    entryPoints: {
      'support/index': 'cypress/support/index.ts',
      ...intEntry,
    },
    outdir: '../../bin/build/integration/cypress',
    logLevel: 'info',
    bundle: true,
    define: { global: 'window' },
    minify: true,
    sourcemap: 'linked',
    target: ['chrome80', 'firefox99', 'safari12', 'edge79'],
    watch: process.argv.includes('--watch'),
  })
  .catch((err) => {
    console.error(err)
    process.exit(1)
  })
