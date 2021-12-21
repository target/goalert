import React, { useState, useEffect } from 'react'
import { Box } from '@mui/system'
import InfiniteScroll from 'react-infinite-scroll-component'
import { DateTime } from 'luxon'
import { DebugMessage } from '../../../schema'
import OutgoingLogCard from './OutgoingLogCard'
import { useFuse } from './hooks'
import { useURLParam } from '../../actions'
import { Typography } from '@mui/material'
import Spinner from '../../loading/components/Spinner'

const LOAD_AMOUNT = 15

interface Props {
  debugMessages?: DebugMessage[]
  onSelect: (debugMessage: DebugMessage) => void
}

export default function OutgoingLogsList(props: Props): JSX.Element {
  const { debugMessages = [], onSelect } = props

  const [searchTerm] = useURLParam('search', '')
  const [start] = useURLParam('start', '')
  const [end] = useURLParam('end', '')

  const [limit, setLimit] = useState(1)

  const { setSearch, results } = useFuse<DebugMessage>({
    data: debugMessages,
    keys: ['status'], // todo: add more keys, phone number/service/user name/etc
    options: { shouldSort: false },
    customOptions: { showResultsWhenNoSearchTerm: true },
  })

  console.log(searchTerm, debugMessages, results)

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
    setSearch(searchTerm)
  }, [searchTerm])

  // what appends stuff to results
  function onNext(): void {
    setLimit(limit + 1)
  }

  function hasMore(): boolean {
    if (filteredResults.length > limit * LOAD_AMOUNT) {
      return true
    }

    return false
  }

  return (
    <InfiniteScroll
      hasMore={hasMore()}
      next={onNext}
      scrollableTarget='content'
      endMessage={
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
      }
      loader={
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            padding: '0.25em 0 0.25em 0',
          }}
        >
          <Spinner text='Loading...' />
        </div>
      }
      dataLength={filteredResults.length}
    >
      <Box
        display='flex'
        flexDirection='column'
        alignItems='stretch'
        width='full'
      >
        {/* TODO: change card's outline color in list when selected */}
        {filteredResults
          .slice(0, limit * LOAD_AMOUNT)
          .map(({ item: debugMessage }) => (
            <OutgoingLogCard
              key={debugMessage.id}
              debugMessage={debugMessage}
              onClick={() => onSelect(debugMessage)}
            />
          ))}
      </Box>
    </InfiniteScroll>
  )
}
