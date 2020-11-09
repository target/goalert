import React from 'react'
import { PropTypes as p } from 'prop-types'
import DialogContent from '@material-ui/core/DialogContent'
import Typography from '@material-ui/core/Typography'
import Error from '@material-ui/icons/Error'
import { styles as globalStyles } from '../../styles/materialStyles'
import { Zoom, makeStyles } from '@material-ui/core'

const useStyles = makeStyles((theme) => ({
  ...globalStyles(theme),
}))
function DialogContentError(props) {
  const classes = useStyles()
  const { error, noPadding, ...other } = props
  const style = noPadding ? { paddingBottom: 0 } : null

  // maintains screen space if no error
  if (!error) {
    return (
      <DialogContent style={style}>
        <Typography
          component='p'
          variant='subtitle1'
          style={{ display: 'flex' }}
        >
          &nbsp;
        </Typography>
      </DialogContent>
    )
  }

  return (
    <DialogContent style={{ textAlign: 'center', ...style }} {...other}>
      <Zoom in>
        <Typography
          component='p'
          variant='subtitle1'
          style={{ display: 'flex' }}
        >
          <Error className={classes.error} />
          &nbsp;
          <span className={classes.error}>{error}</span>
        </Typography>
      </Zoom>
    </DialogContent>
  )
}

DialogContentError.propTypes = {
  error: p.string,
  noPadding: p.bool,
}

export default DialogContentError
