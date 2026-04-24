import React from 'react'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import AdminNumberLookup from './AdminNumberLookup'
import AdminSMSSend from './AdminSMSSend'
import { useConfigValue } from '../util/RequireConfig'

export default function AdminToolbox(): React.JSX.Element {

  const [hasTwilio] = useConfigValue('Twilio.Enable')

  if (!hasTwilio) {
    return (
      <Typography variant='body1' color='textSecondary'>
        Twilio is not enabled. Please enable Twilio in the configuration to use
        this feature.
      </Typography>
    )
  }

  return (
    <Grid
      container
      spacing={2}
      sx={(theme) => ({
        [theme.breakpoints.up('md')]: { justifyContent: 'center' },
      })}
    >
      <Grid size={12} container>
        <Grid size={12}>
          <Typography
            component='h2'
            variant='subtitle1'
            color='textSecondary'
            sx={{ fontSize: '1.1rem' }}
          >
            Twilio Number Lookup
          </Typography>
        </Grid>
        <Grid size={12}>
          <AdminNumberLookup />
        </Grid>
      </Grid>
      <Grid size={12} container>
        <Grid size={12}>
          <Typography
            component='h2'
            variant='subtitle1'
            color='textSecondary'
            sx={{ fontSize: '1.1rem' }}
          >
            Send SMS
          </Typography>
        </Grid>
        <Grid size={12}>
          <AdminSMSSend />
        </Grid>
      </Grid>
    </Grid>
  )
}
