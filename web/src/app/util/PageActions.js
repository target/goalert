import React, { useState, useEffect } from 'react'
import p from 'prop-types'
import { debounce } from 'lodash-es'

const PageActionsContext = React.createContext({
  actions: null,
  setActions: () => {},
})
PageActionsContext.displayName = 'PageActionsContext'

export function PageActionContainer() {
  return (
    <PageActionsContext.Consumer>
      {({ actions }) => actions}
    </PageActionsContext.Consumer>
  )
}

export function PageActionProvider(props) {
  // state = {
  //   actions: null,
  // }
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
        'PageActions: Found more than one <PageActions> component mounted within the same provider.',
      )
    }
  }

  return (
    <PageActionsContext.Provider
      value={{
        actions: actions,
        setActions: handleSetActions,
        trackMount: updateMounted,
      }}
    >
      {props.children}
    </PageActionsContext.Provider>
  )
}

function PageActionUpdater(props) {
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
 * <PageActions>
 *   <Search />
 * </PageActions>
 *
 */
export default function PageActions(props) {
  return (
    <PageActionsContext.Consumer>
      {({ setActions, trackMount }) => (
        <PageActionUpdater
          setActions={setActions}
          trackMount={trackMount}
          children={props.children}
        />
      )}
    </PageActionsContext.Consumer>
  )
}

PageActionUpdater.propTypes = {
  setActions: p.func.isRequired,
}
