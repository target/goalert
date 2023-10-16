import React from 'react'
import {
  DataGrid,
  GridValueGetterParams,
  GridRenderCellParams,
  GridValidRowModel,
  GridToolbar,
} from '@mui/x-data-grid'
import AppLink from '../../util/AppLink'
import {
  EscalationPolicyStep,
  IntegrationKey,
  IntegrationKeyType,
  Notice,
  Service,
  Target,
  TargetType,
} from '../../../schema'
import { Grid, Stack, Tooltip } from '@mui/material'
import CheckCircleOutlineIcon from '@mui/icons-material/CheckCircleOutline'
import {
  ConstructionOutlined,
  ErrorOutline,
  WarningAmberOutlined,
} from '@mui/icons-material'

interface AdminServiceTableProps {
  services: Service[]
  loading: boolean
}

export default function AdminServiceTable(
  props: AdminServiceTableProps,
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
      field: 'status',
      headerName: 'Status',
      width: 150,
      renderCell: (params: GridRenderCellParams<GridValidRowModel>) => {
        const targets: TargetType[] = []
        if (params.row.escalationPolicy.steps?.length) {
          params.row.escalationPolicy.steps?.map(
            (step: EscalationPolicyStep) => {
              step.targets.map((tgt: Target) => {
                if (!targets.includes(tgt.type)) targets.push(tgt.type)
              })
            },
          )
        }

        const noEPSteps = !targets.length
        const noIntegrations =
          !params.row.integrationKeys.length &&
          !params.row.heartbeatMonitors.length
        const hasNotices = !!params.row.notices.length
        const inMaintenance = params.row.maintenanceExpiresAt

        if (!noEPSteps && !noIntegrations && !hasNotices && !inMaintenance)
          return <CheckCircleOutlineIcon color='success' />

        return (
          <Stack direction='row'>
            {noEPSteps && (
              <Tooltip title='Service has empty escalation policy.'>
                <WarningAmberOutlined color='warning' />
              </Tooltip>
            )}
            {noIntegrations && (
              <Tooltip title='Service has no alert integrations configured.'>
                <WarningAmberOutlined color='warning' />
              </Tooltip>
            )}
            {inMaintenance && (
              <Tooltip title='Service is in maintenance mode.'>
                <ConstructionOutlined color='warning' />
              </Tooltip>
            )}
            {hasNotices &&
              params.row.notices.map((notice: Notice, idx: number) => {
                return (
                  <Tooltip key={idx} title={notice.message}>
                    <ErrorOutline color='error' />
                  </Tooltip>
                )
              })}
          </Stack>
        )
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
      field: 'epStepTargets',
      headerName: 'EP Step Target Type(s)',
      width: 215,
      valueGetter: (params: GridValueGetterParams) => {
        const targets: TargetType[] = []
        if (params.row.escalationPolicy.steps?.length) {
          params.row.escalationPolicy.steps?.map(
            (step: EscalationPolicyStep) => {
              step.targets.map((tgt: Target) => {
                if (!targets.includes(tgt.type)) targets.push(tgt.type)
              })
            },
          )
        }
        return targets.sort().join(', ')
      },
    },
    {
      field: 'intKeyTargets',
      headerName: 'Integration Key Target Type(s)',
      width: 250,
      valueGetter: (params: GridValueGetterParams) => {
        const targets: IntegrationKeyType[] = []
        if (params.row.integrationKeys?.length) {
          params.row.integrationKeys?.map((key: IntegrationKey) => {
            if (!targets.includes(key.type)) targets.push(key.type)
          })
        }
        return targets.sort().join(', ')
      },
    },
    {
      field: 'heartbeatMonitors',
      headerName: 'Heartbeat Monitor Total',
      width: 250,
      valueGetter: (params: GridValueGetterParams) => {
        return params.row.heartbeatMonitors?.length
      },
    },
  ]
  return (
    <Grid container sx={{ height: '800px' }}>
      <DataGrid
        rows={services || []}
        loading={loading}
        columns={columns}
        rowSelection={false}
        slots={{ toolbar: GridToolbar }}
        pageSizeOptions={[10, 25, 50, 100]}
      />
    </Grid>
  )
}
