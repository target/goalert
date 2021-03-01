import React from 'react'
import ReactMarkdown from 'react-markdown'
import { safeURL } from './safeURL'
import gfm from 'remark-gfm'
import { makeStyles } from '@material-ui/core'

const useStyles = makeStyles({
  markdown: {
    '& h1, h2, h3, h4, h5, h6, p': {
      marginTop: 0,
      marginBottom: '0.5rem',
    },
    '& hr': {
      margin: '1.5rem 0',
    },
    '& th': {
      textAlign: 'left',
      paddingRight: '8px',
    },
    '& td': {
      paddingRight: '8px',
    },
    '& pre': {
      display: 'block',
      padding: '9.5px',
      margin: '0 0 10px',
      fontSize: '13px',
      lineHeight: '1.42857143',
      color: '#333',
      wordBreak: 'break-all',
      wordWrap: 'break-word',
      backgroundColor: '#f5f5f5',
      border: '1px solid #ccc',
      borderRadius: '4px',
    },
    '& code': {
      padding: '2px 4px',
      fontSize: '90%',
      color: '#c7254e',
      backgroundColor: '#f9f2f4',
      borderRadius: '4px',
    },
    '& pre code': {
      padding: 0,
      fontSize: 'inherit',
      color: 'inherit',
      whiteSpace: 'pre-wrap',
      backgroundColor: 'inherit',
      borderRadius: 0,
    },
  },
})

export default function Markdown(props) {
  const classes = useStyles()
  const { value, ...rest } = props
  if (!value) return null

  return (
    <ReactMarkdown
      className={classes.markdown}
      source={value}
      plugins={[gfm]}
      allowNode={(node) => {
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
