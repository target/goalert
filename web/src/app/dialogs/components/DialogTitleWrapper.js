import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import AppBar from '@material-ui/core/AppBar'
import DialogTitle from '@material-ui/core/DialogTitle'
import IconButton from '@material-ui/core/IconButton'
import Toolbar from '@material-ui/core/Toolbar'
import Typography from '@material-ui/core/Typography'
import withStyles from '@material-ui/core/styles/withStyles'
import CloseIcon from '@material-ui/icons/Close'
import DropDownMenu from '../../dialogs/components/DropDownMenu'
import { styles } from '../../styles/materialStyles'

/**
 * Renders a fullscreen dialog with an app bar if on a small
 * or mobile screen, and a standard dialog title otherwise.
 */
@withStyles(styles)
export default class DialogTitleWrapper extends Component {
  static propTypes = {
    fullScreen: p.bool.isRequired,
    closeIcon: p.object,
    toolbarItems: p.array, // list of JSX items to display on the toolbar
    title: p.string.isRequired,
    onClose: p.func,
    options: p.array, // list of options to display as list items from option icon
  }

  render() {
    const {
      classes,
      closeIcon,
      fullScreen,
      toolbarItems,
      onClose,
      options,
      title,
    } = this.props

    let menu
    if (options && options.length > 0 && fullScreen) {
      menu = <DropDownMenu options={options} color='white' />
    } else if (options && options.length > 0) {
      menu = (
        <div className={classes.topRightActions}>
          <DropDownMenu options={options} />
        </div>
      )
    }

    let closeButton
    if (onClose) {
      closeButton = (
        <IconButton color='inherit' onClick={onClose} aria-label='Close'>
          {closeIcon || <CloseIcon />}
        </IconButton>
      )
    }

    if (fullScreen) {
      return (
        <AppBar position='sticky' style={{ marginBottom: '1em' }}>
          <Toolbar>
            {closeButton}
            <Typography
              component='h2'
              color='inherit'
              style={{ fontSize: '1.2em', flex: 1 }}
            >
              {title}
            </Typography>
            {toolbarItems}
            {menu}
          </Toolbar>
        </AppBar>
      )
    } else {
      return (
        <React.Fragment>
          <DialogTitle disableTypography key='title'>
            <Typography variant='h6' variantMapping={{ h6: 'h2' }}>
              {title}
            </Typography>
          </DialogTitle>
          {menu}
        </React.Fragment>
      )
    }
  }
}
