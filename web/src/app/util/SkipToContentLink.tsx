import React from 'react'
import Link from '@mui/material/Link'
import Typography from '@mui/material/Typography'

export function SkipToContentLink(): JSX.Element {
  return (
    <Link
      sx={{
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
      }}
      href='#content'
    >
      <Typography>Skip to content</Typography>
    </Link>
  )
}
