import React from 'react'
import {
  CardContent,
  CardHeader,
  Typography,
  Card,
  CardActions,
  Chip,
} from '@mui/material'
import { DateTime } from 'luxon'
import { DebugMessage } from '../../../schema'
import toTitleCase from '../../util/toTitleCase'

interface Props {
  debugMessage: DebugMessage
  selected: boolean
  onSelect: () => void
}

export default function DebugMessageCard(props: Props): JSX.Element {
  const { debugMessage, selected, onSelect } = props

  const type = debugMessage.type
  const status = toTitleCase(debugMessage.status)
  const statusDict = {
    success: {
      backgroundColor: '#EDF6ED',
      color: '#1D4620',
    },
    error: {
      backgroundColor: '#FDEBE9',
      color: '#611A15',
    },
    warning: {
      backgroundColor: '#FFF4E5',
      color: '#663C00',
    },
    info: {
      backgroundColor: '#E8F4FD',
      color: '#0E3C61',
    },
  }

  let statusStyles
  const s = status.toLowerCase()
  if (s.includes('deliver')) statusStyles = statusDict.success
  if (s.includes('sent')) statusStyles = statusDict.success
  if (s.includes('fail')) statusStyles = statusDict.error
  if (s.includes('temp')) statusStyles = statusDict.warning
  if (s.includes('pend')) statusStyles = statusDict.info

  return (
    <Card
      onClick={onSelect}
      key={debugMessage.id}
      sx={
        selected
          ? { border: '2px solid green', cursor: 'pointer' }
          : { border: 'none', cursor: 'pointer' }
      }
    >
      <CardHeader
        action={
          <Typography color='textSecondary' sx={{ pr: 1 }}>
            {DateTime.fromISO(debugMessage.createdAt).toFormat('fff')}
          </Typography>
        }
        title={`${type} Notification`}
        titleTypographyProps={{
          color: 'textSecondary',
          gutterBottom: true,
          sx: {
            fontSize: 14,
          },
        }}
        subheader={`Destination: ${debugMessage.destination}`}
        subheaderTypographyProps={{
          component: 'span',
          variant: 'h6',
          color: 'textPrimary',
        }}
        style={{ paddingBottom: 0 }}
      />
      {(debugMessage.serviceName || debugMessage.userName) && (
        <CardContent>
          {debugMessage.serviceName && (
            <Typography color='textSecondary'>
              Service: {debugMessage.serviceName}
            </Typography>
          )}
          {debugMessage.userName && (
            <Typography color='textSecondary'>
              User: {debugMessage.userName}
            </Typography>
          )}
        </CardContent>
      )}
      <CardActions sx={{ p: 2 }}>
        <Chip
          label={status}
          style={statusStyles}
          sx={{
            padding: '2px 8px',
            borderRadius: '15px',
          }}
        />
      </CardActions>
    </Card>
  )
}
