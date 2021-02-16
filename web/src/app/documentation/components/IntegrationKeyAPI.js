import React from 'react'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import Typography from '@material-ui/core/Typography'
import markdownText from '../IntegrationKeys.md'
import Markdown from '../../util/Markdown'
import { useConfigValue } from '../../util/RequireConfig'
import { pathPrefix } from '../../env'

export default function IntegrationKeyAPI() {
  const [publicURL] = useConfigValue('General.PublicURL')

  const finalText = markdownText.replaceAll(
    'https://<example.goalert.me>',
    publicURL || `${window.location.origin}${pathPrefix}`,
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
