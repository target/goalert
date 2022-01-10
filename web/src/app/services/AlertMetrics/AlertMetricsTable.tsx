import React from 'react'
import {
  DataGrid,
  GridRenderCellParams,
  GridValueGetterParams,
  GridValueFormatterParams,
  GridToolbarContainer,
  GridToolbarExport,
  GridToolbarColumnsButton,
  GridToolbarDensitySelector,
  GridToolbarFilterButton,
  gridClasses,
} from '@mui/x-data-grid'
import { Grid } from '@mui/material'
import DownloadIcon from '@mui/icons-material/Download'
import { makeStyles } from '@mui/styles'
import { Alert } from '../../../schema'
import { DateTime } from 'luxon'
import AppLink from '../../util/AppLink'
import AlertMetricsCSV from './AlertMetricsCSV'

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
      width: 250,
      valueFormatter: (params: GridValueFormatterParams) => {
        return `${DateTime.fromISO(params.value as string).toFormat('ccc, DD, t ZZZZ')}`
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
    const getFileName = (): string => {
      if (alerts.length) {
        return 'GoAlert_Alert_Metrics[' + alerts[0].service?.name + ']'
      }
      return 'GoAlert_Alert_Metrics'
    }

    return (
      <GridToolbarContainer className={gridClasses.toolbarContainer}>
        <Grid container>
          <Grid item xs={8}>
            <GridToolbarColumnsButton />
            <GridToolbarFilterButton />
            <GridToolbarDensitySelector />
            <GridToolbarExport
              data-cy='table-metrics-download'
              csvOptions={{
                fileName: getFileName(),
              }}
              printOptions={{ disableToolbarButton: true }}
            />
          </Grid>
          <Grid item xs={4}>
            <AlertMetricsCSV alerts={alerts} />
          </Grid>
        </Grid>
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
            ExportIcon: DownloadIcon,
            Toolbar: CustomToolbar,
          }}
        />
      </Grid>
    </Grid>
  )
}
