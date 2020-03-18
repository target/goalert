import React, { useState, useEffect, useRef } from 'react'
import p from 'prop-types'
import { makeStyles } from '@material-ui/core'
import AppBar from '@material-ui/core/AppBar'
import Hidden from '@material-ui/core/Hidden'
import IconButton from '@material-ui/core/IconButton'
import InputAdornment from '@material-ui/core/InputAdornment'
import Slide from '@material-ui/core/Slide'
import TextField from '@material-ui/core/TextField'
import Toolbar from '@material-ui/core/Toolbar'
import { Close as CloseIcon, Search as SearchIcon } from '@material-ui/icons'
import { DEBOUNCE_DELAY } from '../config'
import AppBarSearchContainer from './AppBarSearchContainer'
import { useURLParam } from '../actions'

const useStyles = makeStyles(theme => {
  return {
    hasSearch: {
      [theme.breakpoints.up('md')]: {
        // type=text added to increase specificity to override the 180 min-width below
        '& input[type=text]': {
          minWidth: 275,
        },
      },
    },
    textField: {
      backgroundColor: 'white',
      borderRadius: '4px',
      [theme.breakpoints.up('md')]: {
        minWidth: 250,
        '& input:focus': {
          minWidth: 275,
        },
        '& input': {
          minWidth: 180,
          transitionProperty: 'min-width',
          transitionDuration: theme.transitions.duration.standard,
          transitionTimingFunction: theme.transitions.easing.easeInOut,
        },
      },
    },
  }
})

/*
 * Renders a search text field that utilizes the URL params to regulate
 * what data to display
 *
 * On a mobile device the the search icon will be present, and when tapped
 * a new appbar will display that contains a search field to use.
 *
 * On a larger screen, the field will always be present to use in the app bar.
 */
export default function Search(props) {
  const [searchParam, setSearchParam] = useURLParam('search', '')

  const classes = useStyles()
  const [search, setSearch] = useState(searchParam)
  const [showMobile, setShowMobile] = useState(Boolean(search))
  const fieldRef = useRef()
  let textClass = classes.textField
  if (search) {
    textClass += ' ' + classes.hasSearch
  }

  // If the page search param changes, we update state directly.
  useEffect(() => {
    setSearch(searchParam)
  }, [searchParam])

  // When typing, we setup a debounce before updating the URL.
  useEffect(() => {
    const t = setTimeout(() => {
      setSearchParam(search)
    }, DEBOUNCE_DELAY)

    return () => clearTimeout(t)
  }, [search])

  function renderTextField(extraProps) {
    return (
      <TextField
        InputProps={{
          ref: fieldRef,
          startAdornment: (
            <InputAdornment position='start'>
              <SearchIcon color='action' />
            </InputAdornment>
          ),
          endAdornment: props.endAdornment && (
            <InputAdornment position='end'>
              {React.cloneElement(props.endAdornment, { anchorRef: fieldRef })}
            </InputAdornment>
          ),
        }}
        data-cy='search-field'
        placeholder='Search'
        variant='outlined'
        margin='dense'
        hiddenLabel
        onChange={e => setSearch(e.target.value)}
        value={search}
        className={textClass}
        {...extraProps}
      />
    )
  }

  function renderMobile() {
    return (
      <AppBarSearchContainer>
        <IconButton
          key='search-icon'
          color='inherit'
          aria-label='Search'
          data-cy='open-search'
          onClick={() => setShowMobile(true)}
        >
          <SearchIcon />
        </IconButton>
        <Slide
          key='search-field'
          in={showMobile || Boolean(search)}
          direction='down'
          mountOnEnter
          unmountOnExit
          style={{
            zIndex: 9001,
          }}
        >
          <AppBar>
            <Toolbar>
              <IconButton
                color='inherit'
                onClick={() => {
                  // cancel search and close the bar
                  setSearch('')
                  setShowMobile(false)
                }}
                aria-label='Cancel'
                data-cy='close-search'
              >
                <CloseIcon />
              </IconButton>
              {renderTextField({ style: { flex: 1 } })}
            </Toolbar>
          </AppBar>
        </Slide>
      </AppBarSearchContainer>
    )
  }

  return (
    <React.Fragment>
      <Hidden smDown>{renderTextField()}</Hidden>
      <Hidden mdUp>{renderMobile()}</Hidden>
    </React.Fragment>
  )
}

Search.propTypes = {
  endAdornment: p.node,
}
