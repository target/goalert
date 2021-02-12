import React, { useContext, useEffect, useState } from 'react'
import p from 'prop-types'
import { debounce } from 'lodash'

const PageActionsContext = React.createContext({
  actions: null,
  setActions: () => {},
})
PageActionsContext.displayName = 'PageActionsContext'

export const PageActionContainer = () => {
  const { actions } = useContext(PageActionsContext)
  return actions
}

export const PageActionProvider = (props) => {
  const [actions, setActions] = useState(null)

  let _mountCount = 0
  let _mounted = false
  let _pending = null

  const _setActions = (actions) => {
    if (!_mounted) {
      _pending = actions
      return
    }

    setActions(actions)
  }

  const debouncedSetActions = debounce(_setActions)

  const updateMounted = (mount) => {
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

  useEffect(() => {
    _mounted = true
    if (_pending) {
      debouncedSetActions(_pending)
      _pending = null
    }
    return () => {
      _mounted = false
      _pending = false
      debouncedSetActions.cancel()
    }
  })

  return (
    <PageActionsContext.Provider
      value={{
        actions: actions,
        setActions: debouncedSetActions,
        trackMount: updateMounted,
      }}
    >
      {props.children}
    </PageActionsContext.Provider>
  )
}

const PageActionUpdater = (props) => {
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
  })

  return _mounted ? props.setActions(props.children) : null
}

PageActionUpdater.propTypes = {
  setActions: p.func.isRequired,
  trackMount: p.func.isRequired,
}

/*
 * Usage:
 *
 * <PageActions>
 *   <Search />
 * </PageActions>
 *
 */
const PageActions = (props) => {
  const { setActions, trackMount } = useContext(PageActionsContext)
  return (
    <PageActionUpdater setActions={setActions} trackMount={trackMount}>
      {props.children}
    </PageActionUpdater>
  )
}

export default PageActions
