import React from 'react'
import ReactMarkdown from 'react-markdown'
import { safeURL } from './safeURL'
import gfm from 'remark-gfm'
import { makeStyles, Typography } from '@material-ui/core'

export const useStyles = makeStyles({
  markdown: {
    '& td, th': {
      textAlign: 'left',
      padding: '0.25rem 1rem',
    },
    '& td:first-child, th:first-child': {
      paddingLeft: 0,
    },
    '& td:last-child, th:last-child': {
      paddingRight: 0,
    },
    '& pre': {
      padding: '0.375rem',
      color: '#333',
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
    <Typography component='div'>
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
    </Typography>
  )
}
