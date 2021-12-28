import React, { useEffect } from 'react'
import { Box } from '@mui/system'
import { DateTime } from 'luxon'
import { DebugMessage } from '../../../schema'
import OutgoingLogCard from './OutgoingLogCard'
import { useFuse } from './hooks'
import { useURLParam } from '../../actions'
import { Typography, Button } from '@mui/material'

export const LOAD_AMOUNT = 10

interface KeyedDebugMessage extends DebugMessage {
  additonalKeys: {
    filteredDestination: string
  }
}

interface Props {
  debugMessages?: KeyedDebugMessage[]
  selectedLog: DebugMessage | null
  onSelect: (debugMessage: DebugMessage) => void
}

export default function OutgoingLogsList(props: Props): JSX.Element {
  const { debugMessages = [], selectedLog, onSelect } = props

  const [searchTerm] = useURLParam('search', '')
  const [start] = useURLParam('start', '')
  const [end] = useURLParam('end', '')
  const [_limit, _setLimit] = useURLParam<string>(
    'limit',
    LOAD_AMOUNT.toString(),
  )

  const setLimit = (newLimit: number): void => _setLimit(newLimit.toString())
  const limit = parseInt(_limit, 10)

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

  let filteredResults = results.slice() // copy results array
  filteredResults = filteredResults.filter((result) => {
    if (!start && !end) return true

    const startDT = DateTime.fromISO(start)
    let endDT = DateTime.fromISO(end)
    const createdAtDT = DateTime.fromISO(result.item.createdAt)

    if (start && !end) {
      endDT = DateTime.now()
    }

    if (createdAtDT > startDT && createdAtDT < endDT) {
      return true
    }

    return false
  })

  // reset page load amount when filters change
  useEffect(() => {
    setLimit(LOAD_AMOUNT)
  }, [searchTerm, start, end])

  // set search within fuse on search change
  useEffect(() => {
    setSearch(searchTerm)
  }, [searchTerm])

  // what appends stuff to results
  function onNext(): void {
    setLimit(limit + LOAD_AMOUNT)
  }

  return (
    <Box
      display='flex'
      flexDirection='column'
      alignItems='stretch'
      width='full'
    >
      {filteredResults.slice(0, limit).map(({ item: debugMessage }) => (
        <OutgoingLogCard
          key={debugMessage.id}
          debugMessage={debugMessage}
          selected={selectedLog?.id === debugMessage.id}
          onSelect={() => onSelect(debugMessage)}
        />
      ))}
      {limit < filteredResults.length ? (
        // load more
        <div
          style={{
            marginTop: '0.5rem',
            marginBottom: '0.5rem',
            display: 'flex',
            justifyContent: 'center',
          }}
        >
          <Button variant='contained' color='primary' onClick={onNext}>
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
