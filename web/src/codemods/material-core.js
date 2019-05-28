export default function transformer(file, api) {
  const j = api.jscodeshift

  return j(file.source)
    .find(j.ImportDeclaration)
    .forEach(path => {
      const pathStr = path.value.source.value

      if (!pathStr.startsWith('material-ui/') && pathStr !== 'material-ui')
        return
      path.value.specifiers.forEach(spec => {
        let newPath = pathStr.replace(
          /^material-ui(.*(?=\/))?/,
          '@material-ui/core',
        )
        if (spec.imported) {
          // handles non-default imports
          // import { CardContent } from 'material-ui/Card' -> import CardContent from '@material-ui/core/CardContent'
          // import { CardContent as asdf } from 'material-ui/Card' -> import asdf from '@material-ui/core/CardContent'
          // import { withStyles } from 'material-ui/styles' -> import withStyles from '@material-ui/core/styles/withStyles'
          newPath = '@material-ui/core/' + spec.imported.name
          if (spec.imported.name === 'withStyles') {
            newPath = '@material-ui/core/styles/withStyles'
          }
          path.insertBefore(
            j.importDeclaration(
              [j.importDefaultSpecifier(spec.local)],
              j.literal(newPath),
            ),
          )
        } else {
          // handles default imports
          // import Card from 'material-ui/Card' -> import Card from '@material-ui/core/Card'
          if (pathStr.includes('/colors/'))
            newPath = newPath.replace('/core/', '/core/colors/')
          if (pathStr.includes('/styles/'))
            newPath = newPath.replace('/core/', '/core/styles/')
          path.insertBefore(j.importDeclaration([spec], j.literal(newPath)))
        }
      })

      path.replace()
    })
    .toSource()
}
