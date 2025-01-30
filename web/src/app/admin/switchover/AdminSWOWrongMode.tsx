import React from 'react'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import DatabaseOff from 'mdi-material-ui/DatabaseOff'

export function AdminSWOWrongMode(): React.JSX.Element {
  return (
    <Grid item container alignItems='center' justifyContent='center'>
      <DatabaseOff color='secondary' style={{ width: '100%', height: 256 }} />
      <Grid item>
        <div style={{ textAlign: 'center' }}>
          <Typography color='secondary' variant='h6' style={{ marginTop: 16 }}>
            Unavailable: Application is not in switchover mode.
            <br />
            <br />
            You must start GoAlert with <code>GOALERT_DB_URL_NEXT</code> or{' '}
            <code>--db-url-next</code> to perform a switchover.
          </Typography>
        </div>
      </Grid>
    </Grid>
  )
}
