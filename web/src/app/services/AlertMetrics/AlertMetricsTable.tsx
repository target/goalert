import React, { useMemo } from 'react'
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
  GridValidRowModel,
  GridColDef,
} from '@mui/x-data-grid'
import { Button, Grid } from '@mui/material'
import DownloadIcon from '@mui/icons-material/Download'
import { makeStyles } from '@mui/styles'
import { Alert, Service } from '../../../schema'
import { DateTime, Duration } from 'luxon'
import AppLink from '../../util/AppLink'
import { useWorker } from '../../worker'
import { pathPrefix } from '../../env'

interface AlertMetricsTableProps {
  alerts: Alert[]
  loading: boolean
  serviceName: string
  startTime: string
  endTime: string
}

const useStyles = makeStyles(() => ({
  tableContent: {
    height: '400px',
  },
}))

const columns: GridColDef[] = [
  {
    field: 'createdAt',
    headerName: 'Created At',
    width: 250,
    valueFormatter: (params: GridValueFormatterParams) =>
      DateTime.fromISO(params.value as string).toFormat('ccc, DD, t ZZZZ'),
  },
  {
    field: 'closedAt',
    headerName: 'Closed At',
    width: 250,
    valueFormatter: (params: GridValueFormatterParams) =>
      DateTime.fromISO(params.value as string).toFormat('ccc, DD, t ZZZZ'),
  },
  {
    field: 'timeToAck',
    headerName: 'Ack Time',
    width: 100,
    valueFormatter: (params: GridValueFormatterParams) =>
      Duration.fromISO(params.value as string).toFormat('hh:mm:ss'),
  },
  {
    field: 'timeToClose',
    headerName: 'Close Time',
    width: 100,
    valueFormatter: (params: GridValueFormatterParams) =>
      Duration.fromISO(params.value as string).toFormat('hh:mm:ss'),
  },
  {
    field: 'alertID',
    headerName: 'Alert ID',
    width: 90,
    renderCell: (params: GridRenderCellParams<GridValidRowModel>) => (
      <AppLink to={`/alerts/${params.row.alertID}`}>{params.value}</AppLink>
    ),
  },
  {
    field: 'escalated',
    headerName: 'Escalated',
    width: 90,
  },
  { field: 'noiseReason', headerName: 'Noise Reason', width: 90 },
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
    width: 300,
  },
  {
    field: 'serviceName',
    headerName: 'Service Name',
    width: 200,
    valueGetter: (params: GridValueGetterParams) => {
      return params.row.service?.name || ''
    },

    renderCell: (params: GridRenderCellParams<{ service: Service }>) => {
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

export default function AlertMetricsTable(
  props: AlertMetricsTableProps,
): React.ReactNode {
  const classes = useStyles()
  const alerts = useMemo(
    () => props.alerts.map((a) => ({ ...a, ...a.metrics })),
    [props.alerts],
  )

  const csvOpts = useMemo(
    () => ({
      urlPrefix: location.origin + pathPrefix,
      alerts,
    }),
    [props.alerts],
  )
  const [csvData] = useWorker('useAlertCSV', csvOpts, '')
  const link = useMemo(
    () => URL.createObjectURL(new Blob([csvData], { type: 'text/csv' })),
    [csvData],
  )

  function CustomToolbar(): React.ReactNode {
    return (
      <GridToolbarContainer className={gridClasses.toolbarContainer}>
        <Grid container justifyContent='space-between'>
          <Grid item>
            <GridToolbarColumnsButton />
            <GridToolbarFilterButton />
            <GridToolbarDensitySelector />
          </Grid>
          <Grid item>
            <AppLink
              to={link}
              download={`${props.serviceName.replace(
                /[^a-z0-9]/gi,
                '_',
              )}-metrics-${props.startTime}-to-${props.endTime}.csv`}
            >
              <Button
                size='small'
                startIcon={<DownloadIcon fontSize='small' />}
              >
                Export
              </Button>
            </AppLink>
          </Grid>
        </Grid>
      </GridToolbarContainer>
    )
  }

  return (
    <Grid container className={classes.tableContent}>
      <Grid item xs={12} data-cy='metrics-table'>
        <DataGrid
          rows={alerts}
          loading={props.loading}
          columns={columns}
          rowSelection={false}
          components={{
            ExportIcon: DownloadIcon,
            Toolbar: CustomToolbar,
          }}
        />
      </Grid>
    </Grid>
  )
}
