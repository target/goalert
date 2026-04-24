import React, { useMemo } from 'react'
import {
  DataGrid,
  GridRenderCellParams,
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

const columns: GridColDef[] = [
  {
    field: 'createdAt',
    headerName: 'Created At',
    width: 250,
    valueFormatter: (value: string) =>
      DateTime.fromISO(value).toFormat('ccc, DD, t ZZZZ'),
  },
  {
    field: 'closedAt',
    headerName: 'Closed At',
    width: 250,
    valueFormatter: (value: string) =>
      DateTime.fromISO(value).toFormat('ccc, DD, t ZZZZ'),
  },
  {
    field: 'timeToAck',
    headerName: 'Ack Time',
    width: 100,
    valueFormatter: (value: string) =>
      Duration.fromISO(value).toFormat('hh:mm:ss'),
  },
  {
    field: 'timeToClose',
    headerName: 'Close Time',
    width: 100,
    valueFormatter: (value: string) =>
      Duration.fromISO(value).toFormat('hh:mm:ss'),
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
    valueFormatter: (value: string) => {
      return value.replace('Status', '')
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
    valueGetter: (_value: unknown, row: Alert) => {
      return row.service?.id || ''
    },
    width: 300,
  },
  {
    field: 'serviceName',
    headerName: 'Service Name',
    width: 200,
    valueGetter: (_value: unknown, row: Alert) => {
      return row.service?.name || ''
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
): JSX.Element {
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

  function CustomToolbar(): JSX.Element {
    return (
      <GridToolbarContainer className={gridClasses.toolbarContainer}>
        <Grid container justifyContent='space-between'>
          <Grid>
            <GridToolbarColumnsButton />
            <GridToolbarFilterButton />
            <GridToolbarDensitySelector />
          </Grid>
          <Grid>
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
    <Grid container style={{ height: '400px' }}>
      <Grid size={12} data-cy='metrics-table'>
        <DataGrid
          rows={alerts}
          loading={props.loading}
          columns={columns}
          rowSelection={false}
          slots={{
            exportIcon: DownloadIcon,
            toolbar: CustomToolbar,
          }}
        />
      </Grid>
    </Grid>
  )
}
