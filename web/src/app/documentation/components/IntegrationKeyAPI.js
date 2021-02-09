import React from 'react'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import Typography from '@material-ui/core/Typography'
import markdownText from '../IntegrationKeys.md'
import Markdown from '../../util/Markdown'

export default function IntegrationKeyAPI() {
  const finalText = markdownText.replaceAll(
    'https://<example.goalert.me>',
    window.location.origin,
  )

  return (
    <Card>
      <CardContent>
        <Typography variant='subtitle1' component='div'>
          <Markdown value={finalText} />
        </Typography>
      </CardContent>
    </Card>
  )
}
