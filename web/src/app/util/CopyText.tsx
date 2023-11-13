import React, { useState } from 'react'
import makeStyles from '@mui/styles/makeStyles'
import copyToClipboard from './copyToClipboard'
import ContentCopy from 'mdi-material-ui/ContentCopy'
import AppLink from './AppLink'
import Tooltip, { TooltipProps } from '@mui/material/Tooltip'

const useStyles = makeStyles({
  copyContainer: {
    alignItems: 'center',
    display: 'flex',
    width: 'fit-content',
    wordBreak: 'break-all',
    cursor: 'pointer',
  },
  icon: {
    paddingRight: 4,
  },
})

interface CopyTextProps {
  placement?: TooltipProps['placement']
  title?: string
  value: string
  asURL?: boolean
}

export default function CopyText(props: CopyTextProps): React.ReactNode {
  const classes = useStyles()
  const [copied, setCopied] = useState(false)

  let content
  if (props.asURL) {
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
        <ContentCopy
          className={props.title ? classes.icon : undefined}
          fontSize='small'
        />
        {props.title}
      </AppLink>
    )
  } else {
    content = (
      <span
        className={classes.copyContainer}
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
        <ContentCopy
          color='primary'
          className={props.title ? classes.icon : undefined}
          fontSize='small'
        />
        {props.title}
      </span>
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
