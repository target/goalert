import React, { useMemo } from 'react'
import {
  DataGrid,
  GridRenderCellParams,
  GridValueGetterParams,
  GridToolbarContainer,
  GridToolbarColumnsButton,
  GridToolbarDensitySelector,
  GridToolbarFilterButton,
  gridClasses,
} from '@mui/x-data-grid'
import { Button, Grid } from '@mui/material'
import DownloadIcon from '@mui/icons-material/Download'
import { makeStyles } from '@mui/styles'
import { Alert } from '../../../schema'
import AppLink from '../../util/AppLink'
import { useWorker } from '../../worker'
import { pathPrefix } from '../../env'

interface AlertCountTableProps {
  alerts: Alert[]
  loading: boolean
  startTime: string
  endTime: string
}

const useStyles = makeStyles(() => ({
  tableContent: {
    height: '400px',
  },
}))

const columns = [
  {
    field: 'serviceName',
    headerName: 'Service Name',
    width: 300,
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
  {
    field: 'total',
    headerName: 'Total',
    width: 150,
  },
  {
    field: 'max',
    headerName: 'Max',
    width: 150,
  },
  {
    field: 'avg',
    headerName: 'Average',
    width: 150,
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
]

export default function AlertCountTable(
  props: AlertCountTableProps,
): JSX.Element {
  const classes = useStyles()
  // const alerts = useMemo(
  //   () => props.alerts.map((a) => ({ ...a, ...a.metrics })),
  //   [props.alerts],
  // )

  const csvOpts = useMemo(
    () => ({
      urlPrefix: location.origin + pathPrefix,
      alerts: props.alerts ?? [],
    }),
    [props.alerts],
  )
  const csvData = useWorker('useAlertCSV', csvOpts, '')
  const link = useMemo(
    () => URL.createObjectURL(new Blob([csvData], { type: 'text/csv' })),
    [csvData],
  )

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
            <AppLink
              to={link}
              download={`all-services-alert-counts-${props.startTime}-to-${props.endTime}.csv`}
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
          rows={props.alerts ?? []}
          loading={props.loading}
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
