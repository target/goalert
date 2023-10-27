import React from 'react'
import ReactMarkdown from 'react-markdown'
import { safeURL } from './safeURL'
import remarkGfm from 'remark-gfm'
import remarkBreaks from 'remark-breaks'
import makeStyles from '@mui/styles/makeStyles'
import AppLink from './AppLink'

const useStyles = makeStyles({
  markdown: {
    overflowWrap: 'break-word',
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

function TableCell({ children, isHeader, align, ...rest }) {
  const content = React.Children.map(children, (c) =>
    ['<br>', '<br/>', '<br />'].includes(c) ? <br /> : c,
  )

  return (
    <td style={{ textAlign: align }} {...rest}>
      {content}
    </td>
  )
}

// Markdown accepts plain text to transform into styled html
// Typically it is wrapped in a <Typography component='div' /> component
export default function Markdown(props) {
  const classes = useStyles()
  const { value, ...rest } = props
  if (!value) return null

  return (
    <ReactMarkdown
      className={classes.markdown}
      components={{
        td: TableCell,
        a: ({ node, inline, className, children, ...props }) => (
          <AppLink to={props.href} newTab {...props}>
            {children}
          </AppLink>
        ),
      }}
      remarkPlugins={[remarkGfm, remarkBreaks]}
      allowElement={(element) => {
        if (
          element.tagName === 'a' &&
          element.children[0].type === 'text' &&
          !safeURL(element.properties.href, element.children[0].value)
        ) {
          element.type = 'text'
          element.value = `[${element.children[0].value}](${element.properties.href})`
          delete element.properties.href
        }

        return true
      }}
      {...rest}
    >
      {value}
    </ReactMarkdown>
  )
}
