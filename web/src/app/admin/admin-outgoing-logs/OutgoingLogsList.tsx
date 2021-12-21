import { Box } from '@mui/system'
import { DebugMessage } from '../../../schema'
import React, { useEffect } from 'react'
import OutgoingLogCard from './OutgoingLogCard'
import { useFuse } from './hooks'
import { useURLParam } from '../../actions'

interface Props {
  debugMessages?: DebugMessage[]
  onSelect: (debugMessage: DebugMessage) => void
}

export default function OutgoingLogsList(props: Props): JSX.Element {
  const { debugMessages = [], onSelect } = props

  const [searchTerm] = useURLParam('search', '')
  // const [start] = useURLParam('start', '')
  // const [end] = useURLParam('end', '')

  const { setSearch, results } = useFuse<DebugMessage>({
    data: debugMessages,
    keys: ['status'],
    options: { shouldSort: false },
    customOptions: { showResultsWhenNoSearchTerm: true },
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
