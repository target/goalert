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
    status: a.status,
    summary: a.summary,
    details: a.details,
    serviceID: a.service?.id,
    serviceName: a.service?.name,
  }))

  const getFileName = (): string => {
    if (props.alerts.length) {
      return (
        'GoAlert_Raw_Alert_Metrics[' + props.alerts[0].service?.name + '].csv'
      )
    }
    return 'GoAlert_Raw_Alert_Metrics.csv'
  }

  return (
    <Grid container>
      <Grid item xs={12}>
        <Typography className={classes.paragraph}>
          <CSVLink
            data={data}
            filename={getFileName()}
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
