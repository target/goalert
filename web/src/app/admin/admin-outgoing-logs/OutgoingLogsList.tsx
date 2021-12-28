import React, { useEffect } from 'react'
import { Box } from '@mui/system'
import { DateTime } from 'luxon'
import { DebugMessage } from '../../../schema'
import OutgoingLogCard from './OutgoingLogCard'
import { useFuse } from './hooks'
import { useURLParam } from '../../actions'
import { Typography, Button } from '@mui/material'

export const LOAD_AMOUNT = 50

interface KeyedDebugMessage extends DebugMessage {
  additonalKeys: {
    filteredDestination: string
  }
}

interface Props {
  debugMessages?: KeyedDebugMessage[]
  selectedLog: DebugMessage | null
  onSelect: (debugMessage: DebugMessage) => void
  onLoadMore: () => void
  onResetLoadMore: () => void
  showingLimit: number
}

export default function OutgoingLogsList(props: Props): JSX.Element {
  const {
    debugMessages = [],
    selectedLog,
    onSelect,
    showingLimit,
    onLoadMore,
    onResetLoadMore,
  } = props

  const [searchTerm] = useURLParam('search', '')
  const [start] = useURLParam('start', '')
  const [end] = useURLParam('end', '')

  const { setSearch, results } = useFuse<KeyedDebugMessage>({
    data: debugMessages,
    keys: [
      'destination',
      'userName',
      'serviceName',
      'status',
      'additionalKeys.filteredDestination',
    ],
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

  // reset page load amount when filters change
  useEffect(() => {
    onResetLoadMore()
  }, [searchTerm, start, end])

  // set search within fuse on search change
  useEffect(() => {
    setSearch(searchTerm)
  }, [searchTerm])

  return (
    <Box
      display='flex'
      flexDirection='column'
      alignItems='stretch'
      width='full'
    >
      {filteredResults.slice(0, showingLimit).map(({ item: debugMessage }) => (
        <OutgoingLogCard
          key={debugMessage.id}
          debugMessage={debugMessage}
          selected={selectedLog?.id === debugMessage.id}
          onSelect={() => onSelect(debugMessage)}
        />
      ))}
      {showingLimit < filteredResults.length ? (
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
