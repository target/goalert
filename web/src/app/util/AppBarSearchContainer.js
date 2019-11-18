import React from 'react'
import p from 'prop-types'
import { debounce } from 'lodash-es'

const SearchContext = React.createContext({
  actions: null,
  setActions: () => {},
})
SearchContext.displayName = 'SearchContext'

export function SearchContainer() {
  return (
    <SearchContext.Consumer>{({ actions }) => actions}</SearchContext.Consumer>
  )
}

export class SearchProvider extends React.PureComponent {
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
        'Search: Found more than one <AppBarSearchContainer> component mounted within the same provider.',
      )
    }
  }

  render() {
    return (
      <SearchContext.Provider
        value={{
          actions: this.state.actions,
          setActions: this.setActions,
          trackMount: this.updateMounted,
        }}
      >
        {this.props.children}
      </SearchContext.Provider>
    )
  }
}

class SearchUpdater extends React.PureComponent {
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
 * <AppBarSearchContainer>
 *   <Search />
 * </AppBarSearchContainer>
 *
 */
export default function AppBarSearchContainer(props) {
  return (
    <SearchContext.Consumer>
      {({ setActions, trackMount }) => (
        <SearchUpdater
          setActions={setActions}
          trackMount={trackMount}
          children={props.children}
        />
      )}
    </SearchContext.Consumer>
  )
}
