import React from 'react'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import Typography from '@material-ui/core/Typography'
import integrationKeys from './sections/IntegrationKeys.md'
import webhooks from './sections/Webhooks.md'
import Markdown from '../util/Markdown'
import { useConfigValue } from '../util/RequireConfig'
import { pathPrefix } from '../env'
import { makeStyles } from '@material-ui/core'

const useStyles = makeStyles({
  mBottom: {
    marginBottom: '3rem',
  },
})

export default function IntegrationKeyAPI(): JSX.Element {
  const [publicURL] = useConfigValue('General.PublicURL')
  const classes = useStyles()

  // NOTE list markdown documents here
  let markdownDocs = [integrationKeys, webhooks]

  markdownDocs = markdownDocs.map((md) =>
    md.replaceAll(
      'https://<example.goalert.me>',
      publicURL || `${window.location.origin}${pathPrefix}`,
    ),
  )

  return (
    <React.Fragment>
      {markdownDocs.map((doc, i) => (
        <Card key={i} className={classes.mBottom}>
          <CardContent>
            <Typography variant='body1' component='div'>
              <Markdown value={doc} />
            </Typography>
          </CardContent>
        </Card>
      ))}
    </React.Fragment>
  )
}
