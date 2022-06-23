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
      {hasMore ? (
        // load more
        <div
          style={{
            marginTop: '0.5rem',
            marginBottom: '0.5rem',
            display: 'flex',
            justifyContent: 'center',
          }}
        >
          <Button variant='contained' onClick={onLoadMore}>
            Load more
          </Button>
        </div>
      ) : (
        // done loading
        <Typography
          color='textSecondary'
          variant='body2'
          sx={{
            display: 'flex',
            justifyContent: 'center',
            padding: '0.25em 0 0.25em 0',
          }}
        >
          Displaying all results.
        </Typography>
      )}
    </Grid>
  )
}
