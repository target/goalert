import React, { useContext, useEffect, useRef, useState } from 'react'
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
  const _mountCount = useRef(0)
  const _mounted = useRef(false)
  const _pending = useRef(null)

  const _setActions = (actions) => {
    if (!_mounted.current) {
      _pending.current = actions
      return
    }

    setActions(actions)
  }

  const debouncedSetActions = debounce(_setActions)

  const updateMounted = (mount) => {
    if (mount) {
      _mountCount.current++
    } else {
      _mountCount.current--
    }

    if (_mountCount.current > 1 && global.console && console.error) {
      console.error(
        'PageActions: Found more than one <PageActions> component mounted within the same provider.',
      )
    }
  }

  useEffect(() => {
    _mounted.current = true
    if (_pending.current) {
      debouncedSetActions(_pending.current)
      _pending.current = false
    }
    return () => {
      _mounted.current = false
      _pending.current = false
      debouncedSetActions.cancel()
    }
  }, [])

  return (
    <PageActionsContext.Provider
      value={{
        actions,
        setActions: debouncedSetActions,
        trackMount: updateMounted,
      }}
    >
      {props.children}
    </PageActionsContext.Provider>
  )
}

const PageActionUpdater = (props) => {
  const _mounted = useRef(false)

  useEffect(() => {
    if (!_mounted.current) {
      _mounted.current = true
      props.trackMount(true)
      props.setActions(props.children)
    }

    return () => {
      _mounted.current = false
      props.trackMount(false)
      props.setActions(null)
    }
  }, [])

  if (_mounted.current) {
    props.setActions(props.children)
  }

  return null
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
