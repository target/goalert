import React from 'react'
import p from 'prop-types'
import SpeedDialAction from '@material-ui/lab/SpeedDialAction'
import SpeedDialIcon from '@material-ui/lab/SpeedDialIcon'
import SpeedDial from '@material-ui/lab/SpeedDial'

export default class ScheduleNewOverrideFAB extends React.PureComponent {
  static propsTypes = {
    label: p.string.isRequired,
    actions: p.arrayOf(
      p.shape({
        icon: p.element.isRequired,
        onClick: p.func.isRequired,
        label: p.string.isRequired,
      }),
    ).isRequired,
  }

  state = {
    open: false,
  }

  shownState = false
  _shownTimeout = -1

  render() {
    if (this.state.open !== this.shownState) {
      clearTimeout(this._shownTimeout)
      this._shownTimeout = setTimeout(() => {
        this.shownState = this.state.open
      }, 350)
    }
    const doToggle = () => this.setState({ open: !this.shownState })
    const doOpen = () => this.setState({ open: true })
    const doClose = () => this.setState({ open: false })
    return (
      <SpeedDial
        ariaLabel={this.props.label}
        ButtonProps={{
          'data-cy': 'page-fab',
        }}
        icon={<SpeedDialIcon />}
        onClick={doToggle}
        onClose={doClose}
        onMouseEnter={doOpen}
        onMouseLeave={doClose}
        open={this.state.open}
        style={{
          position: 'fixed',
          bottom: '1em',
          right: '1em',
          zIndex: 9001,
        }}
      >
        {this.props.actions
          .slice()
          .reverse()
          .map((action, idx) => (
            <SpeedDialAction
              key={idx}
              icon={action.icon}
              tooltipOpen
              tooltipTitle={action.label}
              aria-label={action.label}
              onClick={() => {
                this.setState({ open: false })
                action.onClick()
              }}
            />
          ))}
      </SpeedDial>
    )
  }
}
