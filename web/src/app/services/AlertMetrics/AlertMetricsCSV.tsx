import React from 'react'
import { Grid, Typography } from '@mui/material'
import DownloadIcon from '@mui/icons-material/Download'
import makeStyles from '@mui/styles/makeStyles'
import { CSVLink } from 'react-csv'
import { Alert } from '../../../schema'

interface AlertMetricsCSVProps {
  alerts: Alert[]
}

const useStyles = makeStyles({
  paragraph: {
    display: 'flex',
    justifyContent: 'flex-end',
    margin: '1rem',
  },
  anchor: {
    display: 'flex',
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
            filename='GoAlert_alert_metrics.csv'
            className={classes.anchor}
          >
            <React.Fragment>
              <DownloadIcon /> Export to CSV
            </React.Fragment>
          </CSVLink>
        </Typography>
      </Grid>
    </Grid>
  )
}
