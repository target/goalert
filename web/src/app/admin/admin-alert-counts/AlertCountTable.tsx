import React from 'react'
import {
  DataGrid,
  GridRenderCellParams,
  GridValueGetterParams,
} from '@mui/x-data-grid'
import { Grid } from '@mui/material'
import { makeStyles } from '@mui/styles'
import DownloadIcon from '@mui/icons-material/Download'
import AppLink from '../../util/AppLink'
import { AlertCountSeries } from './useAdminAlertCounts'

interface AlertCountTableProps {
  alertCounts: AlertCountSeries[]
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
      return params.row.serviceName || ''
    },
    renderCell: (params: GridRenderCellParams<string>) => {
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
    width: 150,
    valueGetter: (params: GridValueGetterParams) => {
      const data = params.row.data
      let total = 0
      for (let i = 0; i < data.length; i++) {
        total += data[i].total
      }
      return total
    },
  },
  {
    field: 'max',
    headerName: 'Max',
    width: 150,
    valueGetter: (params: GridValueGetterParams) => {
      const data = params.row.data
      let max = 0
      for (let i = 0; i < data.length; i++) {
        if (data[i].total > max) max = data[i].total
      }
      return max
    },
  },
  {
    field: 'avg',
    headerName: 'Average',
    width: 150,
    valueGetter: (params: GridValueGetterParams) => {
      const data = params.row.data
      let total = 0
      for (let i = 0; i < data.length; i++) {
        total += data[i].total
      }
      return total / data.length
    },
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
  // const alertCounts = useMemo(() => props.alertCounts, [props.alertCounts])

  // const csvOpts = useMemo(
  //   () => ({
  //     urlPrefix: location.origin + pathPrefix,
  //     alertCounts: alertCounts,
  //   }),
  //   [props.alertCounts],
  // )
  // const csvData = useWorker('useAlertCSV', csvOpts, '')
  // const link = useMemo(
  //   () => URL.createObjectURL(new Blob([csvData], { type: 'text/csv' })),
  //   [csvData],
  // )

  // function CustomToolbar(): JSX.Element {
  //   return (
  //     <GridToolbarContainer className={gridClasses.toolbarContainer}>
  //       <Grid container justifyContent='space-between'>
  //         <Grid item>
  //           <GridToolbarColumnsButton />
  //           <GridToolbarFilterButton />
  //           <GridToolbarDensitySelector />
  //         </Grid>
  //         <Grid item>
  //           <AppLink
  //             to={link}
  //             download={`all-services-alert-counts-${props.startTime}-to-${props.endTime}.csv`}
  //           >
  //             <Button
  //               size='small'
  //               startIcon={<DownloadIcon fontSize='small' />}
  //             >
  //               Export
  //             </Button>
  //           </AppLink>
  //         </Grid>
  //       </Grid>
  //     </GridToolbarContainer>
  //   )
  // }

  return (
    <Grid container className={classes.tableContent}>
      <Grid item xs={12} data-cy='alert-count-table'>
        <DataGrid
          rows={props.alertCounts ?? []}
          loading={props.loading}
          columns={columns}
          disableSelectionOnClick
          components={{
            ExportIcon: DownloadIcon,
            // Toolbar: CustomToolbar,
          }}
        />
      </Grid>
    </Grid>
  )
}
