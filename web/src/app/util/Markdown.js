import React from 'react'
import ReactMarkdown from 'react-markdown'
import { safeURL } from './safeURL'
import gfm from 'remark-gfm'

export default function Markdown(props) {
  const { value, ...rest } = props
  if (!value) return null

  return (
    <ReactMarkdown
      className='react-markdown'
      source={value}
      plugins={[gfm]}
      allowNode={(node) => {
        if (node.type !== 'link') return true
        if (node.children[0].type !== 'text') return true // only validate text labels
        if (safeURL(node.url, node.children[0].value)) return true
        if (node.url.startsWith('mailto:')) {
          // do not yield native mailto link, use plain text instead
          node.type = 'text'
          delete node.url
          return true
        }

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
