#!/usr/bin/env node
const path = require('path')

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

require('esbuild')
  .build({
    entryPoints: ['explore/explore.tsx'],
    outdir: 'build/static/',
    logLevel: 'info',
    bundle: true,
    define: {
      // 'process.env.NODE_ENV': '"production"',
      global: 'window',
    },
    minify: true,
    sourcemap: true,
    plugins: [dynamicPublicPathPlugin],
    loader: {
      '.png': 'file',
      '.webp': 'file',
      '.js': 'jsx',
    },
    watch: process.argv.includes('--watch'),
  })
  .catch(() => process.exit(1))
