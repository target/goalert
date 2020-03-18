/* eslint @typescript-eslint/no-var-requires: 0 */
/*
 * This codemod will raise all imports to the top of the file,
 * sort alphabetically, and separate local and node_modules/ imports.
 *
 * Additionally:
 * - make local imports relative
 * - individual import specifiers are sorted
 * - unused imports are removed
 * - duplicate imports are merged
 * - duplicate specifiers are removed
 * - missing material-ui imports are added for JSX components
 */
const path = require('path')
const fs = require('fs')

const appDir = path.resolve(__dirname, '../app')

const isLocal = p => p.source.value.startsWith('.')
export default function transformer(file, api) {
  const j = api.jscodeshift
  const root = j(file.source)
  const dir = path.dirname(file.path) + '/.'

  const imports = root.find(j.ImportDeclaration)

  // make imports relative
  imports.forEach(p => {
    const src = '' + p.node.source.value
    if (src.startsWith('.')) {
      return
    }
    const fullPath = path.resolve(appDir, src + '.js')
    if (fs.existsSync(fullPath)) {
      let newPath = path.relative(dir, fullPath).replace(/\.js$/, '')
      if (!newPath.startsWith('.')) {
        newPath = './' + newPath
      }
      p.node.source.value = newPath
    }
  })

  const nodes = imports.nodes()

  const dedup = {}
  nodes.forEach(n => {
    if (!dedup[n.source.value]) {
      dedup[n.source.value] = n
      return
    }

    dedup[n.source.value].specifiers = dedup[n.source.value].specifiers.concat(
      n.specifiers,
    )
  })

  const deduped = Object.keys(dedup).map(k => dedup[k])
  deduped.forEach(n => {
    const dedup = {}
    n.specifiers.forEach(s => {
      if (!dedup[s.local.name]) {
        dedup[s.local.name] = s
      }
    })
    n.specifiers = Object.keys(dedup)
      .map(k => dedup[k])
      // sort specifiers
      .sort((a, b) => {
        if (!a.imported) return -1
        if (!b.imported) return 1
        if (a.imported.name === b.imported.name) return 0
        return a.imported.name < b.imported.name ? -1 : 1
      })
  })

  // sort imports
  const sorted = deduped.sort((a, b) => {
    if (isLocal(a) !== isLocal(b)) {
      return isLocal(a) ? 1 : -1
    }

    // a few overrides
    if (a.source.value === 'react') return -1
    if (b.source.value === 'react') return 1
    if (a.source.value === 'prop-types') return -1
    if (b.source.value === 'prop-types') return 1

    return a.source.value < b.source.value ? -1 : 1
  })

  sorted.forEach(s => {
    s.loc = null
  })

  root
    .find(j.Statement)
    .at(0)
    .insertBefore(sorted)
  imports.remove()

  return root.toSource({ quote: 'single' })
}
