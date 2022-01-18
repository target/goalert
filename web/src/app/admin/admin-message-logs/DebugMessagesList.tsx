import React, { useEffect, useState } from 'react'
import { Box } from '@mui/system'
import { DateTime } from 'luxon'
import { DebugMessage } from '../../../schema'
import DebugMessageCard from './DebugMessageCard'
import { useFuse } from './useFuse'
import { useURLParam } from '../../actions'
import { Typography, Button } from '@mui/material'

interface KeyedDebugMessage extends DebugMessage {
  additonalKeys?: {
    filteredDestination: string
  }
}

interface Props {
  debugMessages?: KeyedDebugMessage[]
  selectedLog: DebugMessage | null
  onSelect: (debugMessage: DebugMessage) => void
  onLoadMore: () => void
  numRendered: number
}

export default function DebugMessagesList(props: Props): JSX.Element {
  const {
    debugMessages = [],
    selectedLog,
    onSelect,
    numRendered,
    onLoadMore,
  } = props

  const [searchTerm] = useURLParam('search', '')
  const [start] = useURLParam('start', '')
  const [end] = useURLParam('end', '')

  const results = useFuse<KeyedDebugMessage>({
    data: debugMessages,
    keys: [
      'destination',
      'userName',
      'serviceName',
      'status',
      'additionalKeys.filteredDestination',
    ],
    search: searchTerm,
    options: {
      shouldSort: false,
      showResultsWhenNoSearchTerm: true,
      ignoreLocation: true,
      useExtendedSearch: true,
    },
  })

  const startDT = start ? DateTime.fromISO(start) : null
  const endDT = end ? DateTime.fromISO(end) : null

  let filteredResults = results.slice() // copy results array
  filteredResults = filteredResults.filter((result) => {
    const createdAtDT = DateTime.fromISO(result.item.createdAt)
    if (startDT && startDT > createdAtDT) return false
    if (endDT && endDT < createdAtDT) return false
    return true
  })

  return (
    <Box
      data-cy='outgoing-message-list'
      display='flex'
      flexDirection='column'
      alignItems='stretch'
      width='full'
    >
      {filteredResults.slice(0, numRendered).map(({ item: debugMessage }) => (
        <DebugMessageCard
          key={debugMessage.id}
          debugMessage={debugMessage}
          selected={selectedLog?.id === debugMessage.id}
          onSelect={() => onSelect(debugMessage)}
        />
      ))}
      {numRendered < filteredResults.length ? (
        // load more
        <div
          style={{
            marginTop: '0.5rem',
            marginBottom: '0.5rem',
            display: 'flex',
            justifyContent: 'center',
          }}
        >
          <Button variant='contained' color='primary' onClick={onLoadMore}>
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
