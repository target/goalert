import React from 'react'
import Alert from '@mui/material/Alert'
import Card from '@mui/material/Card'
import CardHeader from '@mui/material/CardHeader'
import Tooltip from '@mui/material/Tooltip'
import RemoveIcon from '@mui/icons-material/PlaylistRemove'
import AddIcon from '@mui/icons-material/PlaylistAdd'
import DownIcon from '@mui/icons-material/ArrowDownward'
import { Theme } from '@mui/system'

interface DBVersionProps {
  mainDBVersion: string
  nextDBVersion: string
}

export function AdminSWODBVersionCard(props: {
  data: DBVersionProps
}): React.ReactNode {
  const curVer = props.data.mainDBVersion.split(' on ')
  const nextVer = props.data.nextDBVersion.split(' on ')

  return (
    <Card sx={{ height: '100%' }}>
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          padding: '0 16px 0 16px',
          marginBottom: '16px',
          height: '100%',
        }}
      >
        <CardHeader
          title='DB Diff'
          titleTypographyProps={{ sx: { fontSize: '1.25rem' } }}
        />
        <Tooltip title={curVer[1]}>
          <Alert icon={<RemoveIcon />} severity='warning'>
            From {curVer[0]}
          </Alert>
        </Tooltip>
        <DownIcon
          style={{ flexGrow: 1 }}
          sx={{
            alignSelf: 'center',
            color: (theme: Theme) => theme.palette.primary.main,
          }}
        />
        <Tooltip title={nextVer[1]}>
          <Alert icon={<AddIcon />} severity='success' sx={{ mb: '16px' }}>
            To {nextVer[0]}
          </Alert>
        </Tooltip>
      </div>
    </Card>
  )
}
