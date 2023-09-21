import React from 'react'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Typography from '@mui/material/Typography'
import integrationKeys from './sections/IntegrationKeys.md'
import webhooks from './sections/Webhooks.md'
import Markdown from '../util/Markdown'
import { useConfigValue } from '../util/RequireConfig'
import { pathPrefix } from '../env'
import makeStyles from '@mui/styles/makeStyles'

const useStyles = makeStyles({
  mBottom: {
    marginBottom: '3rem',
  },
})

export default function Documentation(): JSX.Element {
  const [publicURL, webhookEnabled] = useConfigValue(
    'General.PublicURL',
    'Webhook.Enable',
  )
  const classes = useStyles()

  // NOTE list markdown documents here
  let markdownDocs = [integrationKeys]
  if (webhookEnabled) {
    markdownDocs.push(webhooks)
  }

  markdownDocs = markdownDocs.map((md) =>
    md.replaceAll(
      'https://<example.goalert.me>',
      publicURL || `${window.location.origin}${pathPrefix}`,
    ),
  )
  interface H1Props {
    children: React.ReactNode
  }

  return (
    <React.Fragment>
      {markdownDocs.map((doc, i) => (
        <Card key={i} className={classes.mBottom} id={i}>
          <CardContent>
            <Typography variant='body1' component='div'>
              <Markdown value={doc}  components={{
                  h1: (props: H1Props) => {
                  const title = props.children?.toString().replace(' ', '-')
                  return (
                    <h1 id={title}>
                    {title}
                    <a href={`#${title}`}> #</a>
                    </h1>
                  )
                  },
                }}/>
            </Typography>
          </CardContent>
        </Card>
      ))}
    </React.Fragment>
  )
}
