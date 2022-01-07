import React from 'react'
import {
  DataGrid,
  GridRenderCellParams,
  GridValueGetterParams,
  GridValueFormatterParams,
  GridToolbarContainer,
  GridToolbarExport,
  gridClasses,
} from '@mui/x-data-grid'
import { Grid } from '@mui/material'
import { makeStyles } from '@mui/styles'
import { Alert } from '../../../schema'
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
      field: 'createdAt',
      headerName: 'Created At',
      width: 200,
      valueFormatter: (params: GridValueFormatterParams) => {
        return `${DateTime.fromISO(params.value as string).toLocaleString(
          DateTime.DATETIME_SHORT,
        )}`
      },
    },
    {
      field: 'alertID',
      headerName: 'Alert ID',
      width: 90,
      renderCell: (params: GridRenderCellParams<string>) => (
        <AppLink to={`/alerts/${params.row.alertID}`}>{params.value}</AppLink>
      ),
    },
    {
      field: 'summary',
      headerName: 'Summary',
      width: 110,
    },
    {
      field: 'details',
      headerName: 'Details',
      width: 110,
    },
    {
      field: 'status',
      headerName: 'Status',
      width: 200,
      valueFormatter: (params: GridValueFormatterParams) => {
        return (params?.value as string).replace(/Status/, '')
      },
    },
    {
      field: 'serviceID',
      headerName: 'Service ID',
      valueGetter: (params: GridValueGetterParams) => {
        return `${params.row.service.id || ''}`
      },
      hide: true,
    },
    {
      field: 'serviceName',
      headerName: 'Service Name',
      hide: true,
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
  ]

  function CustomToolbar(): JSX.Element {
    return (
      <GridToolbarContainer className={gridClasses.toolbarContainer}>
        <GridToolbarExport
          csvOptions={{ fileName: 'GoAlert_Alert_Metrics', allColumns: true }}
        />
      </GridToolbarContainer>
    )
  }

  return (
    <Grid container className={classes.tableContent}>
      <Grid item xs={12}>
        <DataGrid
          rows={alerts}
          columns={columns}
          disableSelectionOnClick
          components={{
            Toolbar: CustomToolbar,
          }}
        />
      </Grid>
    </Grid>
  )
}
