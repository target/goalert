import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import DialogContent from '@material-ui/core/DialogContent'
import Typography from '@material-ui/core/Typography'
import withStyles from '@material-ui/core/styles/withStyles'
import Error from '@material-ui/icons/Error'
import { styles as globalStyles } from '../../styles/materialStyles'
import { Zoom } from '@material-ui/core'

const styles = theme => {
  return {
    ...globalStyles(theme),
    subtitle: {
      fontSize: '1rem',
      fontFamily: 'Roboto, Helvetica, Arial, sans-serif',
      fontWeight: 400,
      lineHeight: 1.75,
      letterSpacing: '0.00938em',
    },
  }
}

@withStyles(styles)
export default class DialogContentError extends Component {
  static propTypes = {
    error: p.string,
    noPadding: p.bool,
  }

  render() {
    const { classes, error, noPadding, ...other } = this.props
    const style = noPadding ? { paddingBottom: 0 } : null

    // maintains screen space if no error
    if (!error) {
      return (
        <DialogContent style={style}>
          <Typography variant='subtitle1' style={{ display: 'flex' }}>
            &nbsp;
          </Typography>
        </DialogContent>
      )
    }

    return (
      <DialogContent style={{ textAlign: 'center', ...style }} {...other}>
        <Zoom in>
          <Typography className={classes.subtitle} style={{ display: 'flex' }}>
            <Error className={classes.error} />
            &nbsp;
            <span className={classes.error}>{error}</span>
          </Typography>
        </Zoom>
      </DialogContent>
    )
  }
}
