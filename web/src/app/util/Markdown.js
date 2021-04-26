import React from 'react'
import ReactMarkdown from 'react-markdown'
import { safeURL } from './safeURL'
import gfm from 'remark-gfm'
import { makeStyles } from '@material-ui/core'

const useStyles = makeStyles({
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

// Markdown accepts plain text to transform into styled html
// Typically it is wrapped in a <Typography component='div' /> component
export default function Markdown(props) {
  const classes = useStyles()
  const { value, ...rest } = props
  if (!value) return null

  return (
    <ReactMarkdown
      className={classes.markdown}
      plugins={[gfm]}
      allowElement={(element) => {
        if (element.type !== 'link') return true
        if (element.children[0].type !== 'text') return true // only validate text labels
        if (safeURL(element.url, element.children[0].value)) return true

        // unsafe URL, or mismatched label, render as text
        element.type = 'text'
        element.children[0].value = `[${element.children[0].value}](${element.url})`
        delete element.url

        return true
      }}
      {...rest}
    >
      {value}
    </ReactMarkdown>
  )
}
