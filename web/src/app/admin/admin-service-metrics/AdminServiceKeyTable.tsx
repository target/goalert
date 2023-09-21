import React from 'react'
import {
  DataGrid,
  GridValueGetterParams,
  GridRenderCellParams,
  GridValidRowModel,
  GridToolbar,
} from '@mui/x-data-grid'
import AppLink from '../../util/AppLink'
import { Service } from '../../../schema'
import { Grid } from '@mui/material'

interface AdminServiceKeyTableProps {
  services: Service[]
  loading: boolean
}

export default function AdminServiceKeyTable(
  props: AdminServiceKeyTableProps,
): JSX.Element {
  const { services, loading } = props
  const columns = [
    {
      field: 'name',
      headerName: 'Service Name',
      width: 300,
      valueGetter: (params: GridValueGetterParams) => {
        return params.row.name || ''
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
      field: 'integrationKeys',
      headerName: 'Integration Key Total',
      width: 215,
      valueGetter: (params: GridValueGetterParams) => {
        return params.row.integrationKeys?.length || 0
      },
    },
    {
      field: 'onCallUsers',
      headerName: 'On Call Users Total',
      width: 215,
      valueGetter: (params: GridValueGetterParams) => {
        return params.row.onCallUsers?.length || 0
      },
    },
    {
      field: 'escalationPolicy',
      headerName: 'EP Step Total',
      width: 215,
      valueGetter: (params: GridValueGetterParams) => {
        return params.row.escalationPolicy.steps?.length || 0
      },
    },
    {
      field: 'heartbeatMonitors',
      headerName: 'Heartbeat Monitor Total',
      width: 215,
      valueGetter: (params: GridValueGetterParams) => {
        return params.row.heartbeatMonitors.steps?.length || 0
      },
    },
  ]
  return (
    <Grid container sx={{ height: '400px' }}>
      <DataGrid
        rows={services}
        loading={loading}
        autoPageSize
        columns={columns}
        rowSelection={false}
        slots={{ toolbar: GridToolbar }}
      />
    </Grid>
  )
}
