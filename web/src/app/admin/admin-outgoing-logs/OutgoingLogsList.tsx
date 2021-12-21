import { Box } from '@mui/system'
import { DebugMessage } from '../../../schema'
import React, { useEffect } from 'react'
import OutgoingLogCard from './OutgoingLogCard'
import { FilterValues } from './OutgoingLogsFilter'
import { useFuse } from './hooks'

interface Props {
  debugMessages: DebugMessage[]
  onSelect: (debugMessage: DebugMessage) => void
  filter: FilterValues
  searchTerm: string
}

export default function OutgoingLogsList(props: Props): JSX.Element {
  const { debugMessages, onSelect, searchTerm } = props

  const { setSearch, results } = useFuse<DebugMessage>({
    data: debugMessages,
    keys: ['status'],
    // options: { minMatchCharLength: 0 },
  })

  console.log(searchTerm, debugMessages, results)

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
      {/* TODO: change card's outline color in list when selected */}
      {results.map(({ item: debugMessage }) => (
        <OutgoingLogCard
          key={debugMessage.id}
          debugMessage={debugMessage}
          onClick={() => onSelect(debugMessage)}
        />
      ))}
    </Box>
  )
}
