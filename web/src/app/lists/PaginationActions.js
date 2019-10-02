import React from 'react'
import p from 'prop-types'
import { debounce } from 'lodash-es'

const PaginationActionsContext = React.createContext({
  actions: null,
  setActions: () => {},
})
PaginationActionsContext.displayName = 'PaginationActionsContext'

export class PaginationActionsContainer extends React.PureComponent {
  render() {
    return (
      <PaginationActionsContext.Consumer>
        {({ actions }) => actions}
      </PaginationActionsContext.Consumer>
    )
  }
}

export class PaginationActionsProvider extends React.PureComponent {
  state = {
    actions: null,
  }

  _mountCount = 0
  _mounted = false

  componentDidMount() {
    this._mounted = true
    if (this._pending) {
      this.setActions(this._pending)
      this._pending = null
    }
  }

  componentWillUnmount() {
    this._mounted = false
    this._pending = false
    this.setActions.cancel()
  }

  _setActions = actions => {
    if (!this._mounted) {
      this._pending = actions
      return
    }

    this.setState({ actions })
  }
  setActions = debounce(this._setActions)

  updateMounted = mount => {
    if (mount) {
      this._mountCount++
    } else {
      this._mountCount--
    }

    if (this._mountCount > 1 && global.console && console.error) {
      console.error(
        'PaginationActions: Found more than one <PaginationActions> component mounted within the same provider.',
      )
    }
  }

  render() {
    return (
      <PaginationActionsContext.Provider
        value={{
          actions: this.state.actions,
          setActions: this.setActions,
          trackMount: this.updateMounted,
        }}
      >
        {this.props.children}
      </PaginationActionsContext.Provider>
    )
  }
}

class PaginationActionsUpdater extends React.PureComponent {
  static propTypes = {
    setActions: p.func.isRequired,
  }

  _mounted = false

  componentDidMount() {
    this._mounted = true
    this.props.trackMount(true)
    this.props.setActions(this.props.children)
  }

  componentWillUnmount() {
    this._mounted = false
    this.props.trackMount(false)
    this.props.setActions(null)
  }

  render() {
    if (this._mounted) {
      this.props.setActions(this.props.children)
    }

    return null
  }
}

/*
 * Usage:
 *
 * <PaginationActions>
 *   <Search />
 * </PaginationActions>
 *
 */
export default class PaginationActions extends React.PureComponent {
  render() {
    return (
      <PaginationActionsContext.Consumer>
        {({ setActions, trackMount }) => (
          <PaginationActionsUpdater
            setActions={setActions}
            trackMount={trackMount}
            children={this.props.children}
          />
        )}
      </PaginationActionsContext.Consumer>
    )
  }
}
