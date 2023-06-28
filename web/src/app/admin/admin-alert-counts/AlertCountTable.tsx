import React, { useMemo, useEffect } from 'react'
import {
  DataGrid,
  GridRenderCellParams,
  GridValueGetterParams,
  gridPaginatedVisibleSortedGridRowEntriesSelector,
  GridToolbarContainer,
  GridToolbarDensitySelector,
  GridToolbarFilterButton,
  GridToolbarColumnsButton,
  useGridApiContext,
  gridClasses,
  GridValidRowModel,
} from '@mui/x-data-grid'
import { Button, Grid } from '@mui/material'
import { makeStyles } from '@mui/styles'
import DownloadIcon from '@mui/icons-material/Download'
import AppLink from '../../util/AppLink'
import { AlertCountSeries } from './useAdminAlertCounts'
import { useWorker } from '../../worker'
import { pathPrefix } from '../../env'

interface AlertCountTableProps {
  alertCounts: AlertCountSeries[]
  loading: boolean
  startTime: string
  endTime: string
  graphData: AlertCountSeries[]
  setGraphData: (data: AlertCountSeries[]) => void
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
      return params.row.serviceName || ''
    },
    renderCell: (params: GridRenderCellParams<GridValidRowModel>) => {
      if (params.row.id && params.value) {
        return (
          <AppLink to={`/services/${params.row.id}`}>{params.value}</AppLink>
        )
      }
      return ''
    },
  },
  {
    field: 'total',
    headerName: 'Total',
    type: 'number',
    width: 150,
    valueGetter: (params: GridValueGetterParams) => params.row.total,
  },
  {
    field: 'max',
    headerName: 'Max',
    type: 'number',
    width: 150,
    valueGetter: (params: GridValueGetterParams) => params.row.max,
  },
  {
    field: 'avg',
    headerName: 'Average',
    type: 'number',
    width: 150,
    valueGetter: (params: GridValueGetterParams) => params.row.avg,
  },
  {
    field: 'serviceID',
    headerName: 'Service ID',
    valueGetter: (params: GridValueGetterParams) => {
      return params.row.id || ''
    },
    hide: true,
    width: 300,
  },
]

export default function AlertCountTable(
  props: AlertCountTableProps,
): JSX.Element {
  const classes = useStyles()

  const csvOpts = useMemo(
    () => ({
      urlPrefix: location.origin + pathPrefix,
      alertCounts: props.alertCounts,
    }),
    [props.alertCounts],
  )
  const [csvData] = useWorker('useAlertCountCSV', csvOpts, '')
  const link = useMemo(
    () => URL.createObjectURL(new Blob([csvData], { type: 'text/csv' })),
    [csvData],
  )

  function CustomToolbar(): JSX.Element {
    const apiRef = useGridApiContext()
    const currentPage = gridPaginatedVisibleSortedGridRowEntriesSelector(
      apiRef,
    ).map((page) => ({
      serviceName: page.model.serviceName,
      dailyCounts: page.model.dailyCounts,
      id: page.model.id,
      total: page.model.total,
      max: page.model.max,
      avg: page.model.avg,
    }))

    useEffect(() => {
      // only set graphData if currentPage changes
      if (JSON.stringify(props.graphData) !== JSON.stringify(currentPage)) {
        props.setGraphData(currentPage)
      }
    }, [currentPage])

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
      <Grid item xs={12} data-cy='alert-count-table'>
        <DataGrid
          rows={props.alertCounts ?? []}
          loading={props.loading}
          autoPageSize
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
