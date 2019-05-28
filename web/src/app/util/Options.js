import React, { Component } from 'react'
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
import { Mutation } from 'react-apollo'
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
export default class Options extends Component {
  static propTypes = {
    asIcon: p.bool, // render icon button as an icon only
    Icon: p.func, // icon component to override default options menu icon
    iconProps: p.object, // extra properties to supply to IconButton
    anchorProps: p.object, // override props to position menu on desktop
    transformProps: p.object, // override props to position menu on desktop
    options: p.array.isRequired, // [{ disabled: false, text: '', onClick: () => { . . . } }]
    positionRelative: p.bool, // if true, disables the options menu being fixed to the top right
  }

  state = {
    anchorEl: null,
    show: false,
    errorMessage: '',
    showErrorDialog: false,
    showOptions: false,
  }

  handleOpenMenu = event => {
    this.setState({
      anchorEl: event.currentTarget,
      show: true,
    })
  }

  handleCloseMenu = () => {
    this.setState({
      show: false,
    })
  }

  handleShowOptions = bool => {
    this.setState({
      showOptions: bool,
    })
  }

  /*
   * Run mutation and catch any errors,
   */
  onMutationSubmit = (o, mutation) => {
    this.handleCloseMenu()
    this.handleShowOptions(false)
    return mutation({ variables: o.mutation.variables }).catch(error =>
      this.setState({ errorMessage: error.message, showErrorDialog: true }),
    )
  }

  onClick = o => {
    Promise.resolve(o.onClick()).catch(error =>
      this.setState({ errorMessage: error.message, showErrorDialog: true }),
    )
  }

  renderItemMutation = (o, idx, type) => {
    // render list or menu item
    const item = mutation => {
      if (type === 'list') {
        return (
          <ListItem button onClick={() => this.onMutationSubmit(o, mutation)}>
            <ListItemText primary={o.text} />
          </ListItem>
        )
      } else if (type === 'menu') {
        return (
          <MenuItem
            key={idx}
            disabled={o.disabled}
            onClick={() => this.onMutationSubmit(o, mutation)}
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
        {mutation => item(mutation)}
      </Mutation>
    )
  }

  renderIconButton = onClick => {
    const { asIcon, Icon, iconProps } = this.props

    if (asIcon) {
      return (
        <Icon
          aria-label='Other Actions'
          data-cy='other-actions'
          color='inherit'
          onClick={onClick}
          {...iconProps}
        />
      )
    } else {
      return (
        <IconButton
          aria-label='Other Actions'
          data-cy='other-actions'
          color='inherit'
          onClick={onClick}
          {...iconProps}
        >
          {Icon || <OptionsIcon />}
        </IconButton>
      )
    }
  }

  renderMobileOptions() {
    const { options } = this.props

    return (
      <Hidden key='mobile-options' mdUp>
        {this.renderIconButton(() => this.handleShowOptions(true))}
        <SwipeableDrawer
          anchor='bottom'
          disableDiscovery
          disableSwipeToOpen
          open={this.state.showOptions}
          onOpen={() => null}
          onClose={() => this.handleShowOptions(false)}
        >
          <div
            tabIndex={0}
            role='button'
            onClick={() => this.handleShowOptions(false)}
            onKeyDown={() => this.handleShowOptions(false)}
          >
            <List data-cy='mobile-actions'>
              {options.map((o, idx) => {
                // render with mutation form if exists
                if (o.mutation) {
                  return this.renderItemMutation(o, idx, 'list')
                }

                // otherwise render as item with onclick func
                return (
                  <ListItem
                    key={idx}
                    button
                    onClick={() => {
                      this.handleShowOptions(false)
                      this.onClick(o)
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

  renderDesktopOptions() {
    const { options, anchorProps, transformProps } = this.props

    return (
      <Hidden key='desktop-options' smDown>
        {this.renderIconButton(this.handleOpenMenu)}
        <Menu
          anchorEl={() => this.state.anchorEl}
          getContentAnchorEl={null}
          open={!!(this.state.show && this.state.anchorEl)}
          onClose={this.handleCloseMenu}
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
              return this.renderItemMutation(o, idx, 'menu')
            }

            // otherwise render as item with onclick func
            return (
              <MenuItem
                key={idx}
                disabled={o.disabled}
                onClick={() => {
                  this.handleCloseMenu()
                  this.onClick(o)
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

  render() {
    const children = [
      <Dialog
        key='error-dialog'
        open={this.state.showErrorDialog}
        onClose={() => this.setState({ showErrorDialog: false })}
        onExited={() => this.setState({ errorMessage: '' })}
      >
        <DialogTitleWrapper
          fullScreen={isWidthDown('md', this.props.width)}
          title='An error occurred'
        />
        <DialogContentError error={this.state.errorMessage} />
      </Dialog>,
      this.renderDesktopOptions(),
      this.renderMobileOptions(),
    ]

    return children
  }
}
