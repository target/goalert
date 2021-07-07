import React from 'react'
import p from 'prop-types'
import SpeedDial from '../util/SpeedDial'
import { AccountSwitch, AccountMinus, AccountPlus } from 'mdi-material-ui'

export default function ScheduleNewOverrideFAB(props) {
  ScheduleNewOverrideFAB.propsTypes = {
    onClick: p.func.isRequired,
  }

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
