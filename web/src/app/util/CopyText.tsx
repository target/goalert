import React, { useState } from 'react'
import makeStyles from '@mui/styles/makeStyles'
import copyToClipboard from './copyToClipboard'
import ContentCopy from 'mdi-material-ui/ContentCopy'
import AppLink from './AppLink'
import { Typography, Tooltip, TooltipProps } from '@mui/material'

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
  noTypography: {
    display: 'inline',
    textDecorationStyle: 'dotted',
    textUnderlineOffset: '0.25rem',
    textDecorationLine: 'underline',
  },
})

interface CopyTextProps {
  placement?: TooltipProps['placement']
  title?: React.ReactNode
  value: string
  asURL?: boolean
  noTypography?: boolean
}

export default function CopyText(props: CopyTextProps): JSX.Element {
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
  } else if (props.noTypography) {
    content = (
      <span
        className={classes.copyContainer + ' ' + classes.noTypography}
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
  } else {
    content = (
      <Typography
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
      </Typography>
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
