import React from 'react'
import Link from '@material-ui/core/Link'
import { makeStyles } from '@material-ui/core'

const useStyles = makeStyles({
  skipLink: {
    position: 'absolute !important',
    height: '1px',
    width: '1px',
    overflow: 'hidden',
    clip: 'rect(1px, 1px, 1px, 1px)',
    '&:hover, &:focus, &:active': {
      top: '5px',
      left: '5px',
      zIndex: 100000,
      clip: 'auto !important',
      display: 'block',
      width: 'auto',
      height: 'auto',
      padding: '15px 23px 14px',
      fontWeight: 'bold',
      fontSize: '14px',
      textDecoration: 'none',
      lineHeight: 'normal',
      color: '#CD1931',
      backgroundColor: '#ffffff',
      borderRadius: '3px',
      boxShadow: '0 0 2px 2px rgba(0, 0, 0, 0.6)',
    },
  },
})

export function SkipToContentLink(props) {
  const classes = useStyles()
  return (
    <Link className={classes.skipLink} href='#content'>
      Skip to content
    </Link>
  )
}
