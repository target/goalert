import React, { useState, useEffect } from 'react'
import AppBar from '@material-ui/core/AppBar'
import Hidden from '@material-ui/core/Hidden'
import IconButton from '@material-ui/core/IconButton'
import Slide from '@material-ui/core/Slide'
import TextField from '@material-ui/core/TextField'
import Toolbar from '@material-ui/core/Toolbar'
import { Close as CloseIcon, Search as SearchIcon } from '@material-ui/icons'
import { styles } from '../styles/materialStyles'
import { useDispatch, useSelector } from 'react-redux'
import { searchSelector } from '../selectors/url'
import { setURLParam } from '../actions/main'
import { DEBOUNCE_DELAY } from '../config'
import { makeStyles } from '@material-ui/core'

const useStyles = makeStyles(theme => {
  return { searchFieldBox: styles(theme).searchFieldBox }
})

/*
 * Renders a search bar that will fix to the top right of the screen (in the app bar)
 *
 * On a mobile device the the search icon will be present, and when tapped
 * a new appbar will display that contains a search field to use.
 *
 * On a larger screen, the field will always be present to use in the app bar.
 */
export default function Search() {
  const searchParam = useSelector(searchSelector)
  const dispatch = useDispatch()
  const setSearchParam = value => dispatch(setURLParam('search', value))
  const classes = useStyles()
  const [search, setSearch] = useState(searchParam)
  const [showMobile, setShowMobile] = useState(Boolean(search))

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
          disableUnderline: true,
          classes: {
            input: classes.searchFieldBox,
          },
        }}
        placeholder='Search'
        onChange={e => setSearch(e.target.value)}
        value={search}
        {...extraProps}
      />
    )
  }

  function renderMobile() {
    return (
      <React.Fragment>
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
      </React.Fragment>
    )
  }

  return (
    <React.Fragment>
      <Hidden smDown>{renderTextField()}</Hidden>
      <Hidden mdUp>{renderMobile()}</Hidden>
    </React.Fragment>
  )
}
