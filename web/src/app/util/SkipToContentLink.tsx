import React from 'react'
import Link from '@mui/material/Link'
import makeStyles from '@mui/styles/makeStyles'
import Typography from '@mui/material/Typography'

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

export function SkipToContentLink(): React.JSX.Element {
  const classes = useStyles()
  return (
    <Link className={classes.skipLink} href='#content'>
      <Typography>Skip to content</Typography>
    </Link>
  )
}
