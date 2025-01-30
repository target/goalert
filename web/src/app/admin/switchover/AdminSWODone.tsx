import React from 'react'
import { Zoom } from '@mui/material'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import DatabaseCheck from 'mdi-material-ui/DatabaseCheck'
import { TransitionGroup } from 'react-transition-group'

export function AdminSWODone(): React.JSX.Element {
  return (
    <TransitionGroup appear={false}>
      <Zoom in timeout={500}>
        <Grid item container alignItems='center' justifyContent='center'>
          <DatabaseCheck
            color='primary'
            style={{ width: '100%', height: 256 }}
          />
          <Grid item>
            <Typography color='primary' variant='h6' style={{ marginTop: 16 }}>
              DB switchover is complete.
            </Typography>
          </Grid>
        </Grid>
      </Zoom>
    </TransitionGroup>
  )
}
