import React from 'react'
import { PropTypes as p } from 'prop-types'
import SwipeableDrawer from '@material-ui/core/SwipeableDrawer'

export default class MobileSideBar extends React.PureComponent {
  static propTypes = {
    show: p.bool.isRequired,
    onChange: p.func.isRequired,
  }

  render() {
    // disable "discover" swiping open on iOS as it has it defaulted to going back a page
    const iOS = process.browser && /iPad|iPhone|iPod/.test(navigator.userAgent)

    return (
      <SwipeableDrawer
        disableDiscovery={iOS}
        open={this.props.show}
        onOpen={() => this.props.onChange(true)}
        onClose={() => this.props.onChange(false)}
      >
        <div
          tabIndex={0}
          role='button'
          onClick={() => this.props.onChange(false)}
          onKeyDown={() => this.props.onChange(false)}
        >
          {this.props.children}
        </div>
      </SwipeableDrawer>
    )
  }
}
