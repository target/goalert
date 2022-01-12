import React from 'react'
import {
  CardContent,
  CardHeader,
  Typography,
  Card,
  CardActions,
  Chip,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { DateTime } from 'luxon'
import { DebugMessage } from '../../../schema'
import toTitleCase from '../../util/toTitleCase'
import { theme } from '../../mui'

const useStyles = makeStyles<typeof theme>((theme) => ({
  card: {
    marginTop: theme.spacing(0.5),
    marginBottom: theme.spacing(0.5),
    cursor: 'pointer',
  },
  chip: {
    padding: '2px 8px',
    borderRadius: '15px',
  },
  msgType: {
    fontSize: 14,
  },
}))

interface Props {
  debugMessage: DebugMessage
  selected: boolean
  onSelect: () => void
}

export default function DebugMessageCard(props: Props): JSX.Element {
  const { debugMessage, selected, onSelect } = props
  const classes = useStyles()

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
      className={classes.card}
      sx={selected ? { border: '2px solid green' } : { border: 'none' }}
    >
      <CardHeader
        action={
          <Typography color='textSecondary' data-cy='created-at'>
            {DateTime.fromISO(debugMessage.createdAt).toFormat('fff')}
          </Typography>
        }
        title={`${type} Notification`}
        titleTypographyProps={{
          'data-cy': 'type',
          className: classes.msgType,
          color: 'textSecondary',
          gutterBottom: true,
        }}
        subheader={`Destination: ${debugMessage.destination}`}
        subheaderTypographyProps={{
          'data-cy': 'destination',
          component: 'span',
          variant: 'h6',
          color: 'textPrimary',
        }}
        style={{ paddingBottom: 0 }}
      />
      {(debugMessage.serviceName || debugMessage.userName) && (
        <CardContent>
          {debugMessage.serviceName && (
            <Typography color='textSecondary' data-cy='service-name'>
              Service: {debugMessage.serviceName}
            </Typography>
          )}
          {debugMessage.userName && (
            <Typography color='textSecondary' data-cy='user-name'>
              User: {debugMessage.userName}
            </Typography>
          )}
        </CardContent>
      )}
      <CardActions>
        <Chip
          className={classes.chip}
          label={status}
          style={statusStyles}
          data-cy='status'
        />
      </CardActions>
    </Card>
  )
}
