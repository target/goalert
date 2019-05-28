import React from 'react'
import p from 'prop-types'
import CircularProgress from '@material-ui/core/CircularProgress'

import { DEFAULT_SPIN_DELAY_MS, DEFAULT_SPIN_WAIT_MS } from '../../config'

/*
 * Show a loading spinner in the center of the container.
 */
export default class Spinner extends React.PureComponent {
  static propTypes = {
    // Wait `delayMs` milliseconds before rendering a spinner.
    delayMs: p.number,

    // Wait `waitMs` before calling onReady.
    waitMs: p.number,

    // onSpin is called when the spinner starts spinning.
    onSpin: p.func,

    // onReady is called once the spinner has spun for `waitMs`.
    onReady: p.func,
  }

  static defaultProps = {
    delayMs: DEFAULT_SPIN_DELAY_MS,
    waitMs: DEFAULT_SPIN_WAIT_MS,
  }

  state = {
    spin: false,
  }

  componentDidMount() {
    this._spin = setTimeout(() => {
      this._spin = null
      this.setState({ spin: true })
      if (this.props.onSpin) this.props.onSpin()

      if (this.props.waitMs && this.props.onReady) {
        this._spin = setTimeout(this.props.onReady, this.props.waitMs)
      }
    }, this.props.delayMs)
  }
  componentWillUnmount() {
    clearTimeout(this._spin)
  }

  render() {
    if (this.props.delayMs && !this.state.spin) return null

    return (
      <div style={{ position: 'absolute', top: '50%', left: '50%' }}>
        <CircularProgress />
      </div>
    )
  }
}
