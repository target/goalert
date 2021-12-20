import { Box } from '@mui/system'
import { DebugMessage } from '../../../schema'
import React from 'react'
import OutgoingLogCard from './OutgoingLogCard'
import { FilterValues } from './OutgoingLogsFilter'

interface Props {
  debugMessages: DebugMessage[]
  onSelect: (debugMessage: DebugMessage) => void
  filter: FilterValues
  searchTerm: string
}

const OutgoingLogsList = ({ debugMessages, onSelect }: Props): JSX.Element => {
  return (
    <Box
      display='flex'
      flexDirection='column'
      alignItems='stretch'
      width='full'
    >
      {/* TODO: change card's outline color in list when selected */}
      {debugMessages.map((debugMessage: DebugMessage) => (
        <OutgoingLogCard
          key={debugMessage.id}
          debugMessage={debugMessage}
          onClick={() => onSelect(debugMessage)}
        />
      ))}
    </Box>
  )
}

export default OutgoingLogsList
