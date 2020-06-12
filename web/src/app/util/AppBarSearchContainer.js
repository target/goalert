import React, { useState, useEffect } from 'react'
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

export function SearchProvider(props) {
  const [actions, setActions] = useState(null)

  let _mountCount = 0
  let _mounted = false
  let _pending = null

  function _handleSetActions(actns) {
    if (!_mounted) {
      _pending = actns
      return
    }

    setActions(actns)
  }

  const handleSetActions = debounce(_handleSetActions)

  useEffect(() => {
    _mounted = true
    if (_pending) {
      handleSetActions(_pending)
      _pending = null
    }

    return () => {
      _mounted = false
      _pending = false
      handleSetActions.cancel()
    }
  }, [])

  function updateMounted(mount) {
    if (mount) {
      _mountCount++
    } else {
      _mountCount--
    }

    if (_mountCount > 1 && global.console && console.error) {
      console.error(
        'Search: Found more than one <AppBarSearchContainer> component mounted within the same provider.',
      )
    }
  }

  return (
    <SearchContext.Provider
      value={{
        actions: actions,
        setActions: props.setActions,
        trackMount: updateMounted,
      }}
    >
      {props.children}
    </SearchContext.Provider>
  )
}

export function SearchUpdater(props) {
  let _mounted = false

  useEffect(() => {
    _mounted = true
    props.trackMount(true)
    props.setActions(props.children)

    return () => {
      _mounted = false
      props.trackMount(false)
      props.setActions(null)
    }
  }, [])

  if (_mounted) {
    return props.setActions(props.children)
  }

  return null
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

SearchProvider.propTypes = {
  actions: null,
}

SearchUpdater.propTypes = {
  setActions: p.func.isRequired,
}
