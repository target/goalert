import React, { Component } from 'react'
import Avatar from '@material-ui/core/Avatar'

export default class BaseAvatar extends Component {
  constructor(props) {
    super(props)

    this.state = {
      valid: false,
    }
    this._mounted = false
  }

  renderFallback() {
    return null
  }
  srcURL(props = this.props) {
    return ''
  }

  loadSuccess = (a, b, c) => {
    if (!this._mounted) return
    this.setState({ valid: true })
  }

  reset(props = this.props) {
    if (!this._mounted) return
    this.setState({ valid: false })
    if (!this.srcURL(props)) {
      return
    }
    this._image = new window.Image()
    this._image.onload = this.loadSuccess
    this._image.src = this.srcURL(props)
  }

  componentDidMount() {
    this._mounted = true
    this.reset()
  }
  componentWillUnmount() {
    // cleanup handlers on unmount
    this._mounted = false
    if (this._image) {
      this._image.onerror = null
      this._image.onload = null
    }
  }
  componentWillReceiveProps(nextProps) {
    if (this.srcURL(this.props) !== this.srcURL(nextProps)) {
      // reset if it changes
      this.reset(nextProps)
    }
  }

  render() {
    const { userID, fallback, ...props } = this.props

    if (this.state.valid) {
      return <Avatar alt='' src={this._image.src} {...props} />
    }

    return (
      <Avatar alt='' data-cy='avatar-fallback' {...props}>
        {this.renderFallback()}
      </Avatar>
    )
  }
}
