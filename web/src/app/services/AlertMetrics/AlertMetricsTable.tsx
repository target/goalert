import React, { useState, useEffect } from 'react'
import {
  DataGrid,
  GridRenderCellParams,
  GridValueGetterParams,
  GridValueFormatterParams,
  GridToolbarContainer,
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
  loading: boolean
}

const useStyles = makeStyles(() => ({
  tableContent: {
    height: '400px',
  },
}))

const PAGE_SIZE = 10

export default function AlertMetricsTable(
  props: AlertMetricsTableProps,
): JSX.Element {
  const classes = useStyles()
  const [page, setPage] = useState(0)
  const [rowCount, setRowCount] = useState(100)

  useEffect(() => {
    if (props.alerts.length) {
      setRowCount(props.alerts.length)
    }
  }, [page, props.alerts])

  const { alerts, loading } = props

  const columns = [
    {
      field: 'createdAt',
      headerName: 'Created At',
      width: 250,
      valueFormatter: (params: GridValueFormatterParams) => {
        return `${DateTime.fromISO(params.value as string).toFormat(
          'ccc, DD, t ZZZZ',
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
      field: 'status',
      headerName: 'Status',
      width: 160,
      valueFormatter: (params: GridValueFormatterParams) => {
        return (params.value as string).replace('Status', '')
      },
    },
    {
      field: 'summary',
      headerName: 'Summary',
      width: 200,
    },
    {
      field: 'details',
      headerName: 'Details',
      width: 200,
    },
    {
      field: 'serviceID',
      headerName: 'Service ID',
      valueGetter: (params: GridValueGetterParams) => {
        return params.row.service?.id || ''
      },
      hide: true,
      width: 300,
    },
    {
      field: 'serviceName',
      headerName: 'Service Name',
      hide: true,
      width: 200,
      valueGetter: (params: GridValueGetterParams) => {
        return params.row.service?.name || ''
      },
      renderCell: (params: GridRenderCellParams<string>) => {
        if (params.row.service?.id && params.value) {
          return (
            <AppLink to={`/services/${params.row.service.id}`}>
              {params.value}
            </AppLink>
          )
        }
        return ''
      },
    },
  ]

  const getRows = (): Alert[] => {
    return alerts.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE)
  }

  function CustomToolbar(): JSX.Element {
    return (
      <GridToolbarContainer className={gridClasses.toolbarContainer}>
        <Grid container justifyContent='space-between'>
          <Grid item>
            <GridToolbarColumnsButton />
            <GridToolbarFilterButton />
            <GridToolbarDensitySelector />
          </Grid>
          <Grid item>
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
          rows={getRows()}
          loading={loading}
          columns={columns}
          disableSelectionOnClick
          components={{
            ExportIcon: DownloadIcon,
            Toolbar: CustomToolbar,
          }}
          onPageChange={(newPage: number) => setPage(newPage)}
          page={page}
          paginationMode='server'
          pageSize={10}
          rowsPerPageOptions={[10]}
          rowCount={rowCount}
        />
      </Grid>
    </Grid>
  )
}
