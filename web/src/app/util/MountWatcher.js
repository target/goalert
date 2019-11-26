import React from 'react'
import p from 'prop-types'

export default class MountWatcher extends React.PureComponent {
  static propTypes = {
    onMount: p.func,
    onUnmount: p.func,
  }

  static defaultProps = {
    onMount: () => {},
    onUnmount: () => {},
  }

  componentDidMount() {
    this.props.onMount()
  }

  componentWillUnmount() {
    this.props.onUnmount()
  }

  render() {
    return this.props.children
  }
}
