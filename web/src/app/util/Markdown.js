import React from 'react'
import p from 'prop-types'
import ReactMarkdown from 'react-markdown'
import { safeURL } from './safeURL'

export default class Markdown extends React.PureComponent {
  static propTypes = {
    value: p.string,
  }

  render() {
    const { value, ...rest } = this.props
    if (!value) return null

    return (
      <ReactMarkdown
        className='react-markdown'
        source={value}
        allowNode={node => {
          if (node.type !== 'link') return true
          if (node.children[0].type !== 'text') return true // only validate text labels
          if (safeURL(node.url, node.children[0].value)) return true

          // unsafe URL, or mismatched label, render as text
          node.type = 'text'
          node.children[0].value = `[${node.children[0].value}](${node.url})`
          delete node.url

          return true
        }}
        {...rest}
      />
    )
  }
}
