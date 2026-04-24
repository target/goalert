import React from 'react'
import { Button, FormHelperText, Grid, Typography } from '@mui/material'
import OpenInNewIcon from '@mui/icons-material/OpenInNew'
import CopyText from '../../util/CopyText'

interface CalendarSuccessFormProps {
  url: string
}

export default function CalenderSuccessForm({
  url,
}: CalendarSuccessFormProps): React.ReactNode {
  const convertedUrl = url.replace(/^https?:\/\//, 'webcal://')
  return (
    <Grid container spacing={2}>
      <Grid item xs={12} style={{ display: 'flex', justifyContent: 'center' }}>
        <Button
          variant='contained'
          href={convertedUrl}
          target='_blank'
          rel='noopener noreferrer'
        >
          Subscribe
          <OpenInNewIcon fontSize='small' sx={{ ml: 1 }} />
        </Button>
      </Grid>
      <Grid item xs={12}>
        <Typography>
          <CopyText title={url} value={url} placement='bottom' asURL />
        </Typography>
        <FormHelperText>
          Some applications require you copy and paste the URL directly
        </FormHelperText>
      </Grid>
    </Grid>
  )
}
