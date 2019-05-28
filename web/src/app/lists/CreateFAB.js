import React from 'react'
import p from 'prop-types'
import AddIcon from '@material-ui/icons/Add'
import Fab from '@material-ui/core/Fab'

export default class CreateFAB extends React.PureComponent {
  static propTypes = {
    onClick: p.func,
  }

  render() {
    return (
      <Fab
        aria-label='Create New'
        data-cy='page-fab'
        color='primary'
        style={{ position: 'fixed', bottom: '1em', right: '1em' }}
        onClick={this.props.onClick}
      >
        <AddIcon />
      </Fab>
    )
  }
}
