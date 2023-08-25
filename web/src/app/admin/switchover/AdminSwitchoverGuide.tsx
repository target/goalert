import React from 'react'
import swoGuide from '../../../../../docs/switchover.md'

import Markdown from '../../util/Markdown'
import { Typography } from '@mui/material'

export default function AdminSwitchoverGuide(): JSX.Element {
  return (
    <Typography component='div'>
      <Markdown value={swoGuide} />
    </Typography>
  )
}
