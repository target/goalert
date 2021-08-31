import React, { useEffect, useRef, useState } from 'react'
import p from 'prop-types'
import { debounce } from 'lodash'

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

export function SearchProvider(props) {
  const [actions, setActions] = useState(null)

  const mountCount = useRef(0)
  const mounted = useRef(false)
  const pending = useRef(null)

  const _setActions = (actions) => {
    if (!mounted.current) {
      pending.current = actions
      return
    }

    setActions(actions)
  }

  const debouncedSetActions = debounce(_setActions)

  useEffect(() => {
    mounted.current = true
    if (pending.current) {
      debouncedSetActions(pending.current)
      pending.current = null
      // Cleanup function
      return () => {
        mounted.current = false
        pending.current = false
        debouncedSetActions.cancel()
      }
    }
  }, [])

  const updateMounted = (mount) => {
    if (mount) {
      mountCount.current++
    } else {
      mountCount.current--
    }

    if (mountCount.current > 1 && global.console && console.error) {
      console.error(
        'Search: Found more than one <AppBarSearchContainer> component mounted within the same provider.',
      )
    }
  }

  return (
    <SearchContext.Provider
      value={{
        actions: actions,
        setActions: debouncedSetActions,
        trackMount: updateMounted,
      }}
    >
      {props.children}
    </SearchContext.Provider>
  )
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
        <SearchUpdater setActions={setActions} trackMount={trackMount}>
          {props.children}
        </SearchUpdater>
      )}
    </SearchContext.Consumer>
  )
}
