import React from 'react'
import { DataGrid, GridRenderCellParams } from '@mui/x-data-grid'
import { Grid } from '@mui/material'
import { makeStyles } from '@mui/styles'
import { Alert } from '../../../schema'
import _ from 'lodash'
import { DateTime } from 'luxon'
import AppLink from '../../util/AppLink'

interface AlertMetricsTableProps {
  alerts: Alert[]
}

const useStyles = makeStyles(() => ({
  tableContent: {
    height: '400px',
  },
}))

export default function AlertMetricsTable(
  props: AlertMetricsTableProps,
): JSX.Element {
  const classes = useStyles()
  const { alerts } = props

  const rows = _.map(alerts, (alert, idx) => {
    return {
      id: idx,
      alertID: alert.id,
      serviceName: alert.service?.name,
      serviceID: alert.service?.id,
      createdAt: DateTime.fromISO(alert.createdAt).toLocaleString(
        DateTime.DATE_SHORT,
      ),
      summary: alert.summary,
      details: alert.details,
    }
  })

  const columns = [
    {
      field: 'alertID',
      headerName: 'Alert ID',
      width: 90,
      renderCell: (params: GridRenderCellParams<string>) => (
        <AppLink to={`/alerts/${params.row.alertID}`}>{params.value}</AppLink>
      ),
    },
    {
      field: 'serviceName',
      headerName: 'ServiceName',
      width: 150,
      renderCell: (params: GridRenderCellParams<string>) => (
        <AppLink to={`/services/${params.row.serviceID}`}>
          {params.value}
        </AppLink>
      ),
    },
    {
      field: 'createdAt',
      headerName: 'Created At',
      width: 150,
    },
    {
      field: 'summary',
      headerName: 'Summary',
      width: 110,
    },
    {
      field: 'details',
      headerName: 'Details',
    },
  ]

  return (
    <Grid container className={classes.tableContent}>
      <Grid item xs={12}>
        <DataGrid rows={rows} columns={columns} disableSelectionOnClick />
      </Grid>
    </Grid>
  )
}
