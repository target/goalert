import React, { Component } from 'react'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import MDReactComponent from 'markdown-react-js'
import markdownText from '../IntegrationKeys.md'

const handleIterate = (Tag, props, children, level) => {
  if (Tag === 'h2') {
    props = {
      ...props,
      id:
        typeof children[0] === 'string'
          ? children[0].replace(' ', '_')
          : children,
    }
  }

  return <Tag {...props}>{children}</Tag>
}

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
          <MDReactComponent text={finalText} onIterate={handleIterate} />
        </CardContent>
      </Card>
    )
  }
}
