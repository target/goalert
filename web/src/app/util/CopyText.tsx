import React, { useState } from 'react'
import { makeStyles } from '@material-ui/core/styles'
import p from 'prop-types'
import copyToClipboard from './copyToClipboard'
import ContentCopy from 'mdi-material-ui/ContentCopy'
import AppLink from './AppLink'
import Tooltip, { TooltipProps } from '@material-ui/core/Tooltip'

const useStyles = makeStyles({
  copyContainer: {
    alignItems: 'center',
    display: 'flex',
    width: 'fit-content',
    wordBreak: 'break-all',
  },
  icon: {
    paddingRight: 4,
  },
})

interface CopyTextProps {
  placement?: TooltipProps['placement']
  title?: string
  value: string
  textOnly?: boolean
}

export default function CopyText(props: CopyTextProps): JSX.Element {
  const classes = useStyles()
  const [copied, setCopied] = useState(false)

  let content
  if (props.textOnly) {
    content = (
      <span
        role='button'
        tabIndex={0}
        onClick={() => {
          copyToClipboard(props.value)
          setCopied(true)
        }}
        onKeyPress={(e) => {
          if (e.key !== 'Enter') {
            return
          }

          copyToClipboard(props.value)
          setCopied(true)
        }}
      >
        {props.title}
      </span>
    )
  } else {
    content = (
      <AppLink
        className={classes.copyContainer}
        to={props.value}
        onClick={(e) => {
          const tgt = e.currentTarget.href

          e.preventDefault()
          copyToClipboard(tgt.replace(/^mailto:/, ''))
          setCopied(true)
        }}
      >
        <ContentCopy className={classes.icon} fontSize='small' />
        {props.title}
      </AppLink>
    )
  }

  return (
    <Tooltip
      TransitionProps={{ onExited: () => setCopied(false) }}
      title={copied ? 'Copied!' : 'Copy'}
      placement={props.placement || 'right'}
    >
      {content}
    </Tooltip>
  )
}

CopyText.propTypes = {
  placement: p.string,
  title: p.string,
  value: p.string.isRequired,
}
