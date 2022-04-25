#!/usr/bin/env node
/* eslint @typescript-eslint/no-var-requires: 0 */
const path = require('path')
const glob = require('glob')

const dynamicPublicPathPlugin = {
  name: 'prefix-path',
  setup(build) {
    build.onResolve({ filter: /\.(png|webp)$/ }, (args) => {
      const needsPrefix =
        args.kind === 'import-statement' && args.pluginData !== 'dynamic'
      return {
        path: path.resolve(args.resolveDir, args.path),
        namespace: needsPrefix ? 'prefix-path' : 'file',
      }
    })

    build.onLoad({ filter: /\.*/, namespace: 'prefix-path' }, async (args) => {
      return {
        pluginData: 'dynamic',
        contents: `
          import p from ${JSON.stringify(args.path)}
          const prefixPath = pathPrefix + "/static/" + p
          export default prefixPath
        `,
        loader: 'js',
      }
    })
  },
}

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
    define: {
      global: 'window',
    },
    minify: true,
    sourcemap: 'linked',
    plugins: [dynamicPublicPathPlugin],
    target: ['chrome80', 'firefox99', 'safari12', 'edge79'],
    banner: {
      js: `var GOALERT_VERSION=${JSON.stringify(process.env.GOALERT_VERSION)};`,
    },
    loader: {
      '.png': 'file',
      '.webp': 'file',
      '.js': 'jsx',
      '.svg': 'dataurl',
      '.md': 'text',
    },
    watch: process.argv.includes('--watch'),
  })
  .catch((err) => {
    console.error(err)
    process.exit(1)
  })
