import React, { useState, useEffect } from 'react'
import { Box } from '@mui/system'
import { DateTime } from 'luxon'
import { DebugMessage } from '../../../schema'
import OutgoingLogCard from './OutgoingLogCard'
import { useFuse } from './hooks'
import { useURLParam } from '../../actions'
import { Typography, Button } from '@mui/material'

const INITIAL_LIMIT = 1
const LOAD_AMOUNT = 50

interface Props {
  debugMessages?: DebugMessage[]
  selectedLog: DebugMessage | null
  onSelect: (debugMessage: DebugMessage) => void
}

export default function OutgoingLogsList(props: Props): JSX.Element {
  const { debugMessages = [], selectedLog, onSelect } = props

  const [searchTerm] = useURLParam('search', '')
  const [start] = useURLParam('start', '')
  const [end] = useURLParam('end', '')

  const [limit, setLimit] = useState(INITIAL_LIMIT)

  const { setSearch, results } = useFuse<DebugMessage>({
    data: debugMessages,
    keys: ['destination', 'userName', 'serviceName', 'status'],
    options: { shouldSort: false, showResultsWhenNoSearchTerm: true },
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

  useEffect(() => {
    setLimit(INITIAL_LIMIT)
  }, [searchTerm, start, end])

  useEffect(() => {
    setSearch(searchTerm)
  }, [searchTerm])

  // what appends stuff to results
  function onNext(): void {
    setLimit(limit + 1)
  }

  return (
    <Box
      display='flex'
      flexDirection='column'
      alignItems='stretch'
      width='full'
    >
      {filteredResults
        .slice(0, limit * LOAD_AMOUNT)
        .map(({ item: debugMessage }) => (
          <OutgoingLogCard
            key={debugMessage.id}
            debugMessage={debugMessage}
            selected={selectedLog?.id === debugMessage.id}
            onSelect={() => onSelect(debugMessage)}
          />
        ))}
      {limit * LOAD_AMOUNT < filteredResults.length ? (
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
