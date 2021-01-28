import React from 'react'
import { PropTypes as p } from 'prop-types'
import AppBar from '@material-ui/core/AppBar'
import DialogTitle from '@material-ui/core/DialogTitle'
import IconButton from '@material-ui/core/IconButton'
import Toolbar from '@material-ui/core/Toolbar'
import Typography from '@material-ui/core/Typography'
import CloseIcon from '@material-ui/icons/Close'
import { styles as globalStyles } from '../../styles/materialStyles'
import { DialogContent, makeStyles } from '@material-ui/core'
import OtherActions from '../../util/OtherActions'

const useStyles = makeStyles((theme) => {
  const { topRightActions } = globalStyles(theme)

  return {
    appBar: {
      marginBottom: '1em',
    },
    appBarTitle: {
      fontSize: '1.2em',
      flex: 1,
    },
    subtitle: {
      overflowY: 'unset',
      flexGrow: 0,
    },
    topRightActions,
    wideScreenTitle: {
      paddingBottom: 0,
    },
  }
})

/**
 * Renders a fullscreen dialog with an app bar if on a small
 * or mobile screen, and a standard dialog title otherwise.
 */
function DialogTitleWrapper(props) {
  const classes = useStyles()

  const {
    closeIcon,
    fullScreen,
    toolbarItems,
    onClose,
    actions,
    subTitle,
    title,
  } = props

  let menu
  if (actions && actions.length > 0 && fullScreen) {
    menu = <OtherActions actions={actions} color='white' />
  } else if (actions && actions.length > 0) {
    menu = (
      <div className={classes.topRightActions}>
        <OtherActions actions={actions} />
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

  let subtitle
  if (subTitle) {
    subtitle = (
      <DialogContent className={classes.subtitle}>
        {typeof subTitle !== 'string' ? (
          subTitle
        ) : (
          <Typography variant='subtitle1' component='p'>
            {subTitle}
          </Typography>
        )}
      </DialogContent>
    )
  }

  if (fullScreen) {
    return (
      <React.Fragment>
        <AppBar position='sticky' className={classes.appBar}>
          <Toolbar>
            {closeButton}
            {typeof title === 'string' ? (
              <Typography
                data-cy='dialog-title'
                color='inherit'
                className={classes.appBarTitle}
              >
                {title}
              </Typography>
            ) : (
              <div data-cy='dialog-title'>{title}</div>
            )}
            {toolbarItems}
            {menu}
          </Toolbar>
        </AppBar>
        {subtitle}
      </React.Fragment>
    )
  }
  return (
    <React.Fragment>
      <DialogTitle
        className={classes.wideScreenTitle}
        key='title'
        data-cy='dialog-title'
      >
        {title}
      </DialogTitle>
      {subtitle}
      {menu}
    </React.Fragment>
  )
}

DialogTitleWrapper.propTypes = {
  fullScreen: p.bool.isRequired,
  closeIcon: p.object,
  toolbarItems: p.array, // list of JSX items to display on the toolbar
  title: p.node.isRequired,
  subTitle: p.node,
  onClose: p.func,
  actions: p.array, // list of actions to display as list items from option icon
}

export default DialogTitleWrapper
