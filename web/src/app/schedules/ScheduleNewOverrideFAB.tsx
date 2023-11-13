import React from 'react'
import SpeedDial from '../util/SpeedDial'
import { AccountSwitch, AccountMinus, AccountPlus } from 'mdi-material-ui'

export default function ScheduleNewOverrideFAB(props: {
  onClick: (action: string) => void
}): React.ReactNode {
  const actions = [
    {
      label: 'Temporarily Replace a User',
      onClick: () => props.onClick('replace'),
      icon: <AccountSwitch />,
    },
    {
      label: 'Temporarily Remove a User',
      onClick: () => props.onClick('remove'),
      icon: <AccountMinus />,
    },
    {
      label: 'Temporarily Add a User',
      onClick: () => props.onClick('add'),
      icon: <AccountPlus />,
    },
  ]

  return <SpeedDial label='Create New Override' actions={actions} />
}
