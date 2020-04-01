import React from 'react'
import Link from '@material-ui/core/Link'
import { makeStyles } from '@material-ui/core'
import Typography from '@material-ui/core/Typography'

const useStyles = makeStyles({
  skipLink: {
    position: 'static',
    height: '1px',
    width: '1px',
    overflow: 'hidden',
    clip: 'rect(1px, 1px, 1px, 1px)',
    '&:focus, &:active': {
      clip: 'auto !important',
      display: 'block',
      width: 'auto',
      height: 'auto',
      padding: '15px 23px 14px',
      backgroundColor: '#ffffff',
    },
  },
})

export function SkipToContentLink(): JSX.Element {
  const classes = useStyles()
  return (
    <Link className={classes.skipLink} href='#content'>
      <Typography>Skip to content</Typography>
    </Link>
  )
}
