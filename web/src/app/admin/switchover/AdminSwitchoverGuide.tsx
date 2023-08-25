import React from 'react'
import swoGuide from '../../../../../docs/switchover.md'

import Markdown from '../../util/Markdown'
import { Button, Card, CardContent, Typography } from '@mui/material'
import AppLink from '../../util/AppLink'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'

export function AdminSwitchoverGuideButton(): JSX.Element {
  return (
    <Button
      variant='contained'
      endIcon={<OpenInNewIcon />}
      component={AppLink}
      to='/admin/switchover/guide'
      newTab
    >
      Switchover Guide
    </Button>
  )
}

export default function AdminSwitchoverGuide(): JSX.Element {
  return (
    <Card>
      <CardContent>
        <Typography component='div'>
          <Markdown value={swoGuide} />
        </Typography>
      </CardContent>
    </Card>
  )
}
