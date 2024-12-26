import React from 'react'
import { PropTypes as p } from 'prop-types'
import AppBar from '@mui/material/AppBar'
import DialogTitle from '@mui/material/DialogTitle'
import IconButton from '@mui/material/IconButton'
import Toolbar from '@mui/material/Toolbar'
import Typography from '@mui/material/Typography'
import CloseIcon from '@mui/icons-material/Close'
import { styles as globalStyles } from '../../styles/materialStyles'
import { DialogContent } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import OtherActions from '../../util/OtherActions'

const useStyles = makeStyles((theme) => {
  const { topRightActions } = globalStyles(theme)

  return {
    subtitle: {
      overflowY: 'unset',
      flexGrow: 0,
    },
    topRightActions,
  }
})

/**
 * Renders a fullscreen dialog with an app bar if on a small
 * or mobile screen, and a standard dialog title otherwise.
 */
function DialogTitleWrapper(props) {
  const classes = useStyles()

  const {
    closeIcon = <CloseIcon />,
    fullScreen = false,
    toolbarItems = [],
    onClose = undefined,
    actions = [],
    subTitle = '',
    title = '',
  } = props

  let menu
  if (actions.length === 0) {
    menu = null
  } else if (fullScreen) {
    menu = <OtherActions actions={actions} />
  } else {
    menu = (
      <div className={classes.topRightActions}>
        <OtherActions actions={actions} />
      </div>
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
        <AppBar position='sticky' sx={{ mb: 2 }}>
          <Toolbar>
            {onClose && (
              <IconButton
                color='inherit'
                onClick={onClose}
                aria-label='Close'
                sx={{ mr: 2 }}
              >
                {closeIcon}
              </IconButton>
            )}
            {typeof title === 'string' ? (
              <Typography
                data-cy='dialog-title'
                color='inherit'
                variant='h6'
                component='h6'
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
      <DialogTitle key='title' data-cy='dialog-title'>
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
  toolbarItems: p.array, // list of React.JSX items to display on the toolbar
  title: p.node.isRequired,
  subTitle: p.node,
  onClose: p.func,
  actions: p.array, // list of actions to display as list items from option icon
}

export default DialogTitleWrapper
