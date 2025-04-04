import React from 'react'
import ReactMarkdown from 'react-markdown'
import { safeURL } from './safeURL'
import remarkGfm from 'remark-gfm'
import remarkBreaks from 'remark-breaks'
import makeStyles from '@mui/styles/makeStyles'
import AppLink from './AppLink'
import timestampSupport from './Markdown.timestampSupport'

function decodeUnicode(url) {
  return url.replace(/\\u([0-9A-Fa-f]{4})/g, (_, p1) => {
    return String.fromCharCode(parseInt(p1, 16))
  });
}

function decodeHtmlEntities(url) {
  const span = document.createElement('span');
  span.innerHTML = url
  return span.innerText
}

function decodeUrl(url) {
  const decodedUrl = decodeURIComponent(url)
  const unicodeDecodedUrl = decodeUnicode(decodedUrl)
  return decodeHtmlEntities(unicodeDecodedUrl)
}

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
          const decodedURL= decodeUrl(props.href)
          props.href=decodedURL
          <AppLink to={decodedURL} newTab {...props}>
            {children}
          </AppLink>
        ),
      }}
      remarkPlugins={[timestampSupport, remarkGfm, remarkBreaks]}
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
