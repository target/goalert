import React, { useState } from 'react'
import {
  DataGrid,
  GridValueGetterParams,
  GridRenderCellParams,
  GridValidRowModel,
  GridToolbar,
  GridSortItem,
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
  NotificationsOffOutlined,
  UpdateDisabledOutlined,
} from '@mui/icons-material'

interface AdminServiceTableProps {
  services: Service[]
  staleAlertServices: { [serviceName in string]: number }
  loading: boolean
}

export default function AdminServiceTable(
  props: AdminServiceTableProps,
): JSX.Element {
  const { services = [], loading, staleAlertServices } = props
  // Community version of MUI DataGrid only supports sortModels with a single SortItem
  const [sortModel, setSortModel] = useState<GridSortItem[]>([
    {
      field: 'status',
      sort: 'desc',
    },
  ])

  const getServiceStatus = (
    service: Service,
  ): {
    hasEPSteps: boolean
    hasIntegrations: boolean
    hasNotices: boolean
    inMaintenance: boolean
    hasStaleAlerts: boolean
  } => {
    const targets: TargetType[] = []
    if (service.escalationPolicy?.steps?.length) {
      service.escalationPolicy?.steps?.map((step: EscalationPolicyStep) => {
        step.targets.map((tgt: Target) => {
          if (!targets.includes(tgt.type)) targets.push(tgt.type)
        })
      })
    }

    return {
      hasEPSteps: !!targets.length,
      hasIntegrations:
        !(!service.integrationKeys.length && !service.heartbeatMonitors.length),
      hasNotices: !!service.notices.length,
      inMaintenance: !!service.maintenanceExpiresAt,
      hasStaleAlerts: staleAlertServices[service.name] > 0,
    }
  }
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
      valueGetter: (params: GridValueGetterParams) => {
        const {
          hasEPSteps,
          hasIntegrations,
          hasNotices,
          inMaintenance,
          hasStaleAlerts,
        } = getServiceStatus(params.row as Service)
        const warnings = []

        if (
          hasEPSteps &&
          hasIntegrations &&
          !hasNotices &&
          !inMaintenance &&
          !hasStaleAlerts
        )
          return warnings.push('OK')
        if (!hasIntegrations)
          warnings.push('Service Missing Alert Integrations')
        if (!hasEPSteps) warnings.push('Service Has Empty Escalation Policy')
        if (inMaintenance) warnings.push('Service In Maintenance Mode')
        if (hasNotices) warnings.push('Service Reaching Alert Limit')
        if (hasStaleAlerts)
          warnings.push(
            `Service Has ${staleAlertServices[params.row.name]} Stale Alerts`,
          )

        return warnings.join(',')
      },
      renderCell: (params: GridRenderCellParams<GridValidRowModel>) => {
        const {
          hasEPSteps,
          hasIntegrations,
          hasNotices,
          inMaintenance,
          hasStaleAlerts,
        } = getServiceStatus(params.row as Service)

        if (hasEPSteps && hasIntegrations && !hasNotices && !inMaintenance)
          return <CheckCircleOutlineIcon color='success' />

        return (
          <Stack direction='row'>
            {!hasIntegrations && (
              <Tooltip title='Service has no alert integrations configured.'>
                <WarningAmberOutlined color='warning' />
              </Tooltip>
            )}
            {!hasEPSteps && (
              <Tooltip title='Service has empty escalation policy.'>
                <NotificationsOffOutlined color='error' />
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
            {hasStaleAlerts && (
              <Tooltip
                title={`Service has ${
                  staleAlertServices[params.row?.name]
                } stale alerts.`}
              >
                <UpdateDisabledOutlined color='warning' />
              </Tooltip>
            )}
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
            <AppLink
              to={`/escalation-policies/${params.row.escalationPolicy.id}`}
            >
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
        rows={services}
        loading={loading}
        columns={columns}
        rowSelection={false}
        slots={{ toolbar: GridToolbar }}
        pageSizeOptions={[10, 25, 50, 100]}
        sortingOrder={['desc', 'asc']}
        sortModel={sortModel}
        onSortModelChange={(model) => setSortModel(model)}
      />
    </Grid>
  )
}
