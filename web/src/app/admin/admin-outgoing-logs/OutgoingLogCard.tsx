import React from 'react'
import { CardContent, Typography, Card, CardActions } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import { Chip } from '@mui/material'
import { DebugMessage } from './AdminOutgoingLogs'

const useStyles = makeStyles((theme) => ({
  card: {
    margin: theme.spacing(1),
    cursor: 'pointer',
  },
}))

interface AdminOutgoingLogsProps {
  debugMessage: DebugMessage
  onClick?: React.MouseEventHandler<HTMLDivElement>
}

export default function OutgoingLogCard({
  debugMessage,
  onClick,
}: AdminOutgoingLogsProps): JSX.Element {
  const classes = useStyles()

  //   We can do all the logic here to convert dates, phone number, etc to readable format
  const type = debugMessage.type.toUpperCase()
  const formattedUsername = debugMessage.userName || '(Unknown)'
  const formattedDestination = debugMessage.destination
    .replace('(', '')
    .replace(')', '')
  const status = debugMessage.status.toUpperCase()

  return (
    <Card onClick={onClick} key={debugMessage.id} className={classes.card}>
      {/* todo - replace w/ codesandbox card */}
      <CardContent>
        <Typography component='h4'>{type}</Typography>
        <Typography variant='body2' component='div'>
          Destination: {formattedUsername} ({formattedDestination})
        </Typography>
        <Typography variant='body2' color='textSecondary'>
          ID: {debugMessage.id}
        </Typography>
      </CardContent>
      <CardActions>
        <Chip size='small' label={status} />
      </CardActions>
    </Card>
  )
}
