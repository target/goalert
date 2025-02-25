import React, { useEffect } from 'react'
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

export default function Documentation(): React.JSX.Element {
  const [publicURL, webhookEnabled] = useConfigValue(
    'General.PublicURL',
    'Webhook.Enable',
  )
  const classes = useStyles()

  // NOTE list markdown documents here
  let markdownDocs = [{ doc: integrationKeys, id: 'integration-keys' }]
  if (webhookEnabled) {
    markdownDocs.push({ doc: webhooks, id: 'webhooks' })
  }

  markdownDocs = markdownDocs.map((md) => ({
    id: md.id,
    doc: md.doc.replaceAll(
      'https://<example.goalert.me>',
      publicURL || `${window.location.origin}${pathPrefix}`,
    ),
  }))

  // useEffect to ensure that the page scrolls to the correct section after rendering
  useEffect(() => {
    const hash = window.location.hash
    if (!hash) return
    const el = document.getElementById(hash.slice(1))
    if (!el) return

    el.scrollIntoView()
  }, [webhookEnabled, publicURL])

  return (
    <React.Fragment>
      {markdownDocs.map((md, i) => (
        <Card key={i} className={classes.mBottom} id={md.id}>
          <CardContent>
            <Typography variant='body1' component='div'>
              <Markdown value={md.doc} />
            </Typography>
          </CardContent>
        </Card>
      ))}
    </React.Fragment>
  )
}
