import React from 'react'
import p from 'prop-types'
import SpeedDial from '../util/SpeedDial'
import { AccountSwitch, AccountMinus, AccountPlus } from 'mdi-material-ui'

export default class ScheduleNewOverrideFAB extends React.PureComponent {
  static propsTypes = {
    onClick: p.func.isRequired,
  }

  actions = [
    {
      label: 'Temporarily Replace a User',
      onClick: () => this.props.onClick('replace'),
      icon: <AccountSwitch />,
    },
    {
      label: 'Temporarily Remove a User',
      onClick: () => this.props.onClick('remove'),
      icon: <AccountMinus />,
    },
    {
      label: 'Temporarily Add a User',
      onClick: () => this.props.onClick('add'),
      icon: <AccountPlus />,
    },
  ]

  render() {
    return <SpeedDial label='Create New Override' actions={this.actions} />
  }
}
