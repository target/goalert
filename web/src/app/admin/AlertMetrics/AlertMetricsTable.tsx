import React from 'react'
import {
  DataGrid,
  GridRenderCellParams,
  GridValueGetterParams,
  GridValueFormatterParams,
} from '@mui/x-data-grid'
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
      valueGetter: (params: GridValueGetterParams) => {
        return `${params.row.service.name || ''}`
      },
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
      valueFormatter: (params: GridValueFormatterParams) => {
        return `${DateTime.fromISO(params.value as string).toLocaleString(
          DateTime.DATE_MED,
        )}`
      },
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
        <DataGrid rows={alerts} columns={columns} disableSelectionOnClick />
      </Grid>
    </Grid>
  )
}
