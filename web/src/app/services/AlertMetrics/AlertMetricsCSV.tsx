import React from 'react'
import { Button } from '@mui/material'
import DownloadIcon from '@mui/icons-material/Download'
import { CSVLink } from 'react-csv'
import { Alert } from '../../../schema'
import { DateTime } from 'luxon'

interface AlertMetricsCSVProps {
  alerts: Alert[]
}

export default function AlertMetricsCSV(
  props: AlertMetricsCSVProps,
): JSX.Element {
  const zoneAbbr = DateTime.local().toFormat('ZZZZ Z')

  // Note: the data object is ordered
  const data = props.alerts.map((a) => ({
    [`createdAt (${zoneAbbr})`]: DateTime.fromISO(a.createdAt).toLocal().toSQL({
      includeOffset: false,
    }),
    [`closedAt (${zoneAbbr})`]: DateTime.fromISO(a.metrics?.closedAt as string)
      .toLocal()
      .toSQL({
        includeOffset: false,
      }),
    alertID: a.alertID,
    status: a.status.replace('Status', ''),
    summary: a.summary,
    details: a.details,
    serviceID: a.service?.id,
    serviceName: a.service?.name,
  }))

  const getFileName = (): string => {
    if (props.alerts.length) {
      return 'GoAlert_Alert_Metrics[' + props.alerts[0].service?.name + '].csv'
    }
    return 'GoAlert_Alert_Metrics.csv'
  }

  return (
    <CSVLink data={data} filename={getFileName()}>
      <Button size='small' startIcon={<DownloadIcon fontSize='small' />}>
        Export
      </Button>
    </CSVLink>
  )
}
