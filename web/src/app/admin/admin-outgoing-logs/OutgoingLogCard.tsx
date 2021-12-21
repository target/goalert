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
  serviceName: {
    marginBottom: 12,
  },
}))

interface AdminOutgoingLogsProps {
  debugMessage: DebugMessage
  index: number
  onClick?: React.MouseEventHandler<HTMLDivElement>
}

export default function OutgoingLogCard({
  debugMessage,
  index,
  onClick,
}: AdminOutgoingLogsProps): JSX.Element {
  const classes = useStyles()

  //   We can do all the logic here to convert dates, phone number, etc to readable format
  const type = debugMessage.type
  const formattedUsername = debugMessage.userName || '(Unknown)'
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
  if (s.includes('fail')) statusStyles = statusDict.error
  if (s.includes('temp')) statusStyles = statusDict.warning
  if (s.includes('pend')) statusStyles = statusDict.info

  return (
    <Card onClick={onClick} key={debugMessage.id} className={classes.card}>
      <CardHeader
        action={
          <Typography color='textSecondary'>
            {index} - {DateTime.fromISO(debugMessage.createdAt).toFormat('fff')}
          </Typography>
        }
        title={`${type} Notification`}
        titleTypographyProps={{
          className: classes.msgType,
          color: 'textSecondary',
          gutterBottom: true,
        }}
        subheader={`To ${formattedUsername}`}
        subheaderTypographyProps={{
          component: 'span',
          variant: 'h6',
          color: 'textPrimary',
        }}
        style={{ paddingBottom: 0 }}
      />
      <CardContent style={{ paddingTop: 0 }}>
        <Typography className={classes.serviceName} color='textSecondary'>
          Service: {debugMessage.serviceName}
        </Typography>
        <Typography variant='body2' component='p'>
          Destination: {debugMessage.destination}
        </Typography>
      </CardContent>
      <CardActions>
        <Chip className={classes.chip} label={status} style={statusStyles} />
      </CardActions>
    </Card>
  )
}
