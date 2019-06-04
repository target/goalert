import React, { Component } from 'react'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import Typography from '@material-ui/core/Typography'
import markdownText from '../IntegrationKeys.md'
import Markdown from '../../util/Markdown'

function replaceAll(target, search, replacement) {
  return target.split(search).join(replacement)
}

const replaceString = 'https://<example.goalert.me>'

export default class IntegrationKeyAPI extends Component {
  render() {
    const protocol = window.location.protocol || 'https:'
    const host = window.location.host

    let finalText = markdownText
    if (host) {
      finalText = replaceAll(finalText, replaceString, protocol + '//' + host)
    }

    return (
      <Card>
        <CardContent>
          <Typography name='details' variant='subtitle1'>
            <Markdown value={finalText} />
          </Typography>
        </CardContent>
      </Card>
    )
  }
}
