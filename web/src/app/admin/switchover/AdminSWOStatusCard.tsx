import React, { useEffect, useState } from 'react'
import { useTheme, SvgIconProps } from '@mui/material'
import ButtonGroup from '@mui/material/ButtonGroup'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CardHeader from '@mui/material/CardHeader'
import Typography from '@mui/material/Typography'
import ResetIcon from 'mdi-material-ui/DatabaseRefresh'
import NoExecuteIcon from 'mdi-material-ui/DatabaseExportOutline'
import ExecuteIcon from 'mdi-material-ui/DatabaseExport'
import ErrorIcon from 'mdi-material-ui/DatabaseAlert'
import IdleIcon from 'mdi-material-ui/DatabaseSettings'
import InProgressIcon from 'mdi-material-ui/DatabaseEdit'
import { SWOStatus } from '../../../schema'
import LoadingButton from '@mui/lab/LoadingButton'
import { toTitle } from './util'
import { AdminSwitchoverGuideButton } from './AdminSwitchoverGuide'

function getIcon(data: SWOStatus): JSX.Element {
  const i: SvgIconProps = { color: 'primary', sx: { fontSize: '3.5rem' } }

  if (data.lastError) {
    return <ErrorIcon {...i} color='error' />
  }

  if (data.state === 'idle') {
    return <IdleIcon {...i} />
  }

  return <InProgressIcon {...i} />
}

function getSubheader(data: SWOStatus): React.ReactNode {
  if (data.lastError) return 'Error'
  if (data.state === 'done') return 'Complete'
  if (data.state === 'idle') return 'Ready'
  if (data.state === 'unknown') return 'Needs Reset'
  return 'Busy'
}

function getDetails(data: SWOStatus): React.ReactNode {
  if (data.lastError) {
    return (
      <Typography color='error' sx={{ pb: 2 }}>
        {toTitle(data.lastError)}
      </Typography>
    )
  }
  if (data?.state !== 'unknown' && data.lastStatus) {
    return <Typography sx={{ pb: 2 }}>{toTitle(data.lastStatus)}</Typography>
  }
  return <Typography>&nbsp;</Typography> // reserves whitespace
}

type AdminSWOStatusCardProps = {
  data: SWOStatus

  onResetClick: () => void
  onExecClick: () => boolean
}

export function AdminSWOStatusCard(
  props: AdminSWOStatusCardProps,
): JSX.Element {
  const theme = useTheme()

  // We track this separately so we can wait for a NEW status without
  // our button flickering back to idle.
  const [state, setState] = useState(props.data.state)
  useEffect(() => {
    setState(props.data.state)
  }, [props.data.state])

  const isExec = ['syncing', 'pausing', 'executing'].includes(state)

  return (
    <Card sx={{ height: '100%' }}>
      <CardContent
        sx={{ display: 'flex', flexDirection: 'column', height: '100%' }}
      >
        <CardHeader
          title='Switchover Status'
          avatar={getIcon(props.data)}
          subheader={getSubheader(props.data)}
          titleTypographyProps={{ sx: { fontSize: '1.25rem' } }}
          sx={{ p: 0 }}
        />
        {getDetails(props.data)}
        <div style={{ flexGrow: 1 }} />
        <ButtonGroup
          orientation={theme.breakpoints.up('md') ? 'vertical' : 'horizontal'}
          sx={{ width: '100%', pb: '32px' }}
        >
          <LoadingButton
            startIcon={<ResetIcon />}
            // disabled={mutationStatus.fetching}
            variant='outlined'
            size='large'
            loading={state === 'resetting'}
            loadingPosition='start'
            onClick={() => {
              setState('resetting')
              props.onResetClick()
            }}
          >
            {state === 'resetting' ? 'Resetting...' : 'Reset'}
          </LoadingButton>
          <LoadingButton
            startIcon={state !== 'idle' ? <NoExecuteIcon /> : <ExecuteIcon />}
            disabled={state !== 'idle'}
            variant='outlined'
            size='large'
            loading={isExec}
            loadingPosition='start'
            onClick={() => {
              if (props.onExecClick()) setState('syncing')
            }}
          >
            {isExec ? 'Executing...' : 'Execute'}
          </LoadingButton>
        </ButtonGroup>
        <AdminSwitchoverGuideButton />
      </CardContent>
    </Card>
  )
}
