#!/usr/bin/env node
const path = require('path')

require('esbuild')
  .build({
    entryPoints: {
      'user-sim': path.resolve(__dirname, './user-sim.ts'),
    },
    outdir: 'bin',
    logLevel: 'info',
    bundle: true,
    sourcemap: 'linked',
    external: ['k6/http'],
    target: ['es6'],
    format: 'esm',
  })
  .catch((err) => {
    console.error(err)
    process.exit(1)
  })
