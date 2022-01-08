import React from 'react'
import { Grid, Typography, Button } from '@mui/material'
import DownloadIcon from '@mui/icons-material/Download'
import makeStyles from '@mui/styles/makeStyles'
import { CSVLink } from 'react-csv'
import { Alert } from '../../../schema'
import { theme } from '../../mui'

interface AlertMetricsCSVProps {
  alerts: Alert[]
}

const useStyles = makeStyles({
  paragraph: {
    display: 'flex',
    justifyContent: 'flex-end',
  },
  anchor: {
    color: theme.palette.primary.main,
    '&:hover': {
      textDecoration: 'none',
    },
  },
})

export default function AlertMetricsCSV(
  props: AlertMetricsCSVProps,
): JSX.Element {
  const classes = useStyles()
  // Note: the data object is ordered
  const data = props.alerts.map((a) => ({
    createdAt: a.createdAt,
    alertID: a.alertID,
    summary: a.summary,
    details: a.details,
    status: a.status,
    serviceID: a.service?.id,
    serviceName: a.service?.name,
  }))

  return (
    <Grid container>
      <Grid item xs={12}>
        <Typography className={classes.paragraph}>
          <CSVLink
            data={data}
            filename='GoAlert_Raw_Alert_Metrics.csv'
            className={classes.anchor}
          >
            <Button data-cy='raw-metrics-download' size='small'>
              <DownloadIcon sx={{ fontSize: '18px', marginRight: '8px' }} />{' '}
              Export Raw
            </Button>
          </CSVLink>
        </Typography>
      </Grid>
    </Grid>
  )
}
