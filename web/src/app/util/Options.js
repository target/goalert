import React, { Component , useState} from 'react'
import { PropTypes as p } from 'prop-types'
import Hidden from '@material-ui/core/Hidden'
import IconButton from '@material-ui/core/IconButton'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import Menu from '@material-ui/core/Menu'
import MenuItem from '@material-ui/core/MenuItem'
import SwipeableDrawer from '@material-ui/core/SwipeableDrawer'
import withStyles from '@material-ui/core/styles/withStyles'
import { MoreHoriz as OptionsIcon } from '@material-ui/icons'
import { styles } from '../styles/materialStyles'
import { Mutation } from '@apollo/client/react/components'
import Dialog from '@material-ui/core/Dialog'
import DialogTitleWrapper from '../dialogs/components/DialogTitleWrapper'
import DialogContentError from '../dialogs/components/DialogContentError'
import withWidth, { isWidthDown } from '@material-ui/core/withWidth/index'

/*
 * Renders options that will fix to the top right of the screen (in the app bar)
 *
 * On a mobile device the menu will slide a drawer in from the bottom of the
 * screen, rendering all options available.
 *
 * On a larger screen, the menu will open a popover menu at the location of
 * the options menu when clicked.
 */
@withWidth()
@withStyles(styles)
export default function Options(props) {

  const [state, setState] = useState({
    anchorEl: null,
    show: false,
    errorMessage: '',
    showErrorDialog: false,
    showOptions: false,
  })

  const handleOpenMenu = (event) => {
    setState({
      anchorEl: event.currentTarget,
      show: true,
    })
  }

  const handleCloseMenu = () => {
    setState({
      show: false,
    })
  }

  const handleShowOptions = (bool) => {
    setState({
      showOptions: bool,
    })
  }

  /*
   * Run mutation and catch any errors,
   */
  const onMutationSubmit = (o, mutation) => {
    handleCloseMenu()
    handleShowOptions(false)
    return mutation({ variables: o.mutation.variables }).catch((error) =>
      setState({ errorMessage: error.message, showErrorDialog: true }),
    )
  }

  const onClick = (o) => {
    Promise.resolve(o.onClick()).catch((error) =>
    setState({ errorMessage: error.message, showErrorDialog: true }),
    )
  }

  const renderItemMutation = (o, idx, type) => {
    // render list or menu item
    const item = (mutation) => {
      if (type === 'list') {
        return (
          <ListItem button onClick={() => onMutationSubmit(o, mutation)}>
            <ListItemText primary={o.text} />
          </ListItem>
        )
      }
      if (type === 'menu') {
        return (
          <MenuItem
            key={idx}
            disabled={o.disabled}
            onClick={() => onMutationSubmit(o, mutation)}
          >
            {o.text}
          </MenuItem>
        )
      }
    }

    // wrap with mutation component
    return (
      <Mutation
        key={idx}
        mutation={o.mutation.query}
        update={(cache, { data }) => {
          // invoke on success function if exists
          if (typeof o.onSuccess === 'function') {
            o.onSuccess(cache, data)
          }
        }}
      >
        {(mutation) => item(mutation)}
      </Mutation>
    )
  }

  const renderIconButton = (onClick) => {
    const { asIcon, Icon, iconProps } = props

    if (asIcon) {
      return (
        <Icon
          aria-label='Other Actions'
          data-cy='other-actions'
          color='inherit'
          onClick={onClick}
          aria-expanded={state.show || state.showOptions}
          {...iconProps}
        />
      )
    }
    return (
      <IconButton
        aria-label='Other Actions'
        data-cy='other-actions'
        color='inherit'
        onClick={onClick}
        aria-expanded={state.show || state.showOptions}
        {...iconProps}
      >
        {Icon || <OptionsIcon />}
      </IconButton>
    )
  }

  function renderMobileOptions() {
    const { options } = props

    return (
      <Hidden key='mobile-options' mdUp>
        {renderIconButton(() => handleShowOptions(true))}
        <SwipeableDrawer
          anchor='bottom'
          disableDiscovery
          disableSwipeToOpen
          open={state.showOptions}
          onOpen={() => null}
          onClose={() => handleShowOptions(false)}
        >
          <div
            tabIndex={0}
            role='button'
            onClick={() => handleShowOptions(false)}
            onKeyDown={() => handleShowOptions(false)}
          >
            <List data-cy='mobile-actions'>
              {options.map((o, idx) => {
                // render with mutation form if exists
                if (o.mutation) {
                  return renderItemMutation(o, idx, 'list')
                }

                // otherwise render as item with onclick func
                return (
                  <ListItem
                    key={idx}
                    button
                    onClick={() => {
                      handleShowOptions(false)
                      onClick(o)
                    }}
                  >
                    <ListItemText primary={o.text} />
                  </ListItem>
                )
              })}
            </List>
          </div>
        </SwipeableDrawer>
      </Hidden>
    )
  }

  function renderDesktopOptions() {
    const { options, anchorProps, transformProps } = props

    return (
      <Hidden key='desktop-options' smDown>
        {renderIconButton(handleOpenMenu)}
        <Menu
          anchorEl={() => state.anchorEl}
          getContentAnchorEl={null}
          open={!!(state.show && state.anchorEl)}
          onClose={handleCloseMenu}
          PaperProps={{
            style: {
              minWidth: '15em',
            },
          }}
          anchorOrigin={
            anchorProps || {
              vertical: 'bottom',
              horizontal: 'right',
            }
          }
          transformOrigin={
            transformProps || {
              vertical: 'top',
              horizontal: 'right',
            }
          }
        >
          {options.map((o, idx) => {
            // render with mutation form if exists
            if (o.mutation) {
              return renderItemMutation(o, idx, 'menu')
            }

            // otherwise render as item with onclick func
            return (
              <MenuItem
                key={idx}
                disabled={o.disabled}
                onClick={() => {
                  handleCloseMenu()
                  onClick(o)
                }}
              >
                {o.text}
              </MenuItem>
            )
          })}
        </Menu>
      </Hidden>
    )
  }
    const children = [
      <Dialog
        key='error-dialog'
        open={state.showErrorDialog}
        onClose={() => setState({ showErrorDialog: false })}
        onExited={() => setState({ errorMessage: '' })}
      >
        <DialogTitleWrapper
          fullScreen={isWidthDown('md', props.width)}
          title='An error occurred'
        />
        <DialogContentError error={state.errorMessage} />
      </Dialog>,
      renderDesktopOptions(),
      renderMobileOptions(),
    ]

    return children
}


Options.propTypes = {
    asIcon: p.bool, // render icon button as an icon only
    Icon: p.func, // icon component to override default options menu icon
    iconProps: p.object, // extra properties to supply to IconButton
    anchorProps: p.object, // override props to position menu on desktop
    transformProps: p.object, // override props to position menu on desktop
    options: p.array.isRequired, // [{ disabled: false, text: '', onClick: () => { . . . } }]
    positionRelative: p.bool, // if true, disables the options menu being fixed to the top right
  }
