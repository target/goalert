import React, { useState, useEffect, useRef } from 'react'
import p from 'prop-types'
import AppBar from '@mui/material/AppBar'
import Hidden from '@mui/material/Hidden'
import IconButton from '@mui/material/IconButton'
import InputAdornment from '@mui/material/InputAdornment'
import Slide from '@mui/material/Slide'
import TextField from '@mui/material/TextField'
import Toolbar from '@mui/material/Toolbar'
import { Close as CloseIcon, Search as SearchIcon } from '@mui/icons-material'
import { DEBOUNCE_DELAY } from '../config'
import AppBarSearchContainer from './AppBarSearchContainer'
import { useURLParam } from '../actions'
import { transitionStyles } from './Transitions'

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
  // track the last value so we know if it changed externally
  // or from a local event so we don't lose typed characters.
  const [prevParamValue, setPrevParamValue] = useState(searchParam)

  const classes = transitionStyles()
  const [search, setSearch] = useState(searchParam)
  const [showMobile, setShowMobile] = useState(Boolean(search))
  const fieldRef = useRef()

  useEffect(() => {
    if (prevParamValue !== searchParam) {
      setSearch(searchParam)
    }
  }, [searchParam])

  // When typing, we setup a debounce before updating the URL.
  useEffect(() => {
    const t = setTimeout(() => {
      setSearchParam(search)
      setPrevParamValue(search)
    }, DEBOUNCE_DELAY)

    return () => clearTimeout(t)
  }, [search])

  function renderTextField() {
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
        name='search'
        hiddenLabel
        onChange={(e) => setSearch(e.target.value)}
        value={search}
        size='small'
        fullWidth={props.fullWidth}
        className={props.transition ? classes.transition : null}
        sx={(theme) => ({
          minWidth: 250,
          backgroundColor: theme.palette.mode === 'dark' ? 'inherit' : 'white',
          borderRadius: '4px',
        })}
      />
    )
  }

  function renderMobile() {
    return (
      <React.Fragment>
        <AppBarSearchContainer>
          <IconButton
            aria-label='Search'
            data-cy='open-search'
            onClick={() => setShowMobile(true)}
            size='large'
            sx={(theme) => ({
              color: theme.palette.mode === 'light' ? 'inherit' : undefined,
            })}
          >
            <SearchIcon />
          </IconButton>
        </AppBarSearchContainer>
        <Slide
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
                onClick={() => {
                  // cancel search and close the bar
                  setSearch('')
                  setShowMobile(false)
                }}
                aria-label='Cancel'
                data-cy='close-search'
                size='large'
                sx={(theme) => ({
                  color: theme.palette.mode === 'light' ? 'inherit' : undefined,
                })}
              >
                <CloseIcon />
              </IconButton>
              {renderTextField()}
            </Toolbar>
          </AppBar>
        </Slide>
      </React.Fragment>
    )
  }

  return (
    <React.Fragment>
      <Hidden mdDown>{renderTextField()}</Hidden>
      <Hidden mdUp>{renderMobile()}</Hidden>
    </React.Fragment>
  )
}

Search.propTypes = {
  endAdornment: p.node,
  transition: p.bool,
  fullWidth: p.bool,
}

Search.defaultProps = {
  transition: true,
}
