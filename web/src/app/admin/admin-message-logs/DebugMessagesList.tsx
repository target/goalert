import React from 'react'
import { DebugMessage } from '../../../schema'
import DebugMessageCard from './DebugMessageCard'
import { Button, Grid, Typography } from '@mui/material'

interface Props {
  debugMessages: DebugMessage[]
  selectedLog: DebugMessage | null
  onSelect: (debugMessage: DebugMessage) => void
  onLoadMore: () => void
  hasMore: boolean
}

export default function DebugMessagesList(props: Props): JSX.Element {
  const { debugMessages, selectedLog, onSelect, hasMore, onLoadMore } = props

  return (
    <Grid container direction='column' spacing={2}>
      {debugMessages.map((msg) => (
        <Grid key={msg.id} item xs={12}>
          <DebugMessageCard
            debugMessage={msg}
            selected={selectedLog?.id === msg.id}
            onSelect={() => onSelect(msg)}
          />
        </Grid>
      ))}

      <Grid
        item
        xs={12}
        sx={{
          display: 'flex',
          justifyContent: 'center',
        }}
      >
        {hasMore ? (
          <Button variant='contained' onClick={onLoadMore}>
            Show more
          </Button>
        ) : (
          <Typography color='textSecondary' variant='body2'>
            Displaying all results.
          </Typography>
        )}
      </Grid>
    </Grid>
  )
}
