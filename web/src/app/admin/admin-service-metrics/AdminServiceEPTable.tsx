import React from 'react'
import {
  DataGrid,
  GridValueGetterParams,
  GridRenderCellParams,
  GridValidRowModel,
  GridToolbar,
} from '@mui/x-data-grid'
import AppLink from '../../util/AppLink'
import { EscalationPolicyStep, Service, TargetType } from '../../../schema'
import { Grid } from '@mui/material'

interface AdminServiceEPTableProps {
  services: Service[]
  loading: boolean
}

export default function AdminServiceEPTable(
  props: AdminServiceEPTableProps,
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
      field: 'epPolicyName',
      headerName: 'Escalation Policy Name',
      width: 300,
      valueGetter: (params: GridValueGetterParams) => {
        return params.row.escalationPolicy?.name || ''
      },
      renderCell: (params: GridRenderCellParams<GridValidRowModel>) => {
        if (
          params.row.escalationPolicy.id &&
          params.row.escalationPolicy.name
        ) {
          return (
            <AppLink to={`/services/${params.row.escalationPolicy.id}`}>
              {params.row.escalationPolicy.name}
            </AppLink>
          )
        }
        return ''
      },
    },
    {
      field: 'epStepTotal',
      headerName: 'EP Step Total',
      width: 150,
      valueGetter: (params: GridValueGetterParams) => {
        return params.row.escalationPolicy.steps?.length || 0
      },
    },
    {
      field: 'epStepTargets',
      headerName: 'EP Step Target Type(s)',
      width: 215,
      valueGetter: (params: GridValueGetterParams) => {
        const targets: TargetType[] = []
        if (params.row.escalationPolicy.steps?.length) {
          params.row.escalationPolicy.steps?.map(
            (step: EscalationPolicyStep) => {
              step.targets.map((tgt) => {
                if (!targets.includes(tgt.type)) targets.push(tgt.type)
              })
            },
          )
        }
        return targets.sort().join(', ')
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
