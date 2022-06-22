import React from 'react'
import { Box } from '@mui/system'
import { DebugMessage } from '../../../schema'
import DebugMessageCard from './DebugMessageCard'
import { Typography, Button } from '@mui/material'

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
    <Box
      data-cy='outgoing-message-list'
      display='flex'
      flexDirection='column'
      alignItems='stretch'
      width='full'
    >
      {debugMessages.map((msg) => (
        <DebugMessageCard
          key={msg.id}
          debugMessage={msg}
          selected={selectedLog?.id === msg.id}
          onSelect={() => onSelect(msg)}
        />
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
    </Box>
  )
}
