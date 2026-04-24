import React from 'react'
import { PropTypes as p } from 'prop-types'
import DialogContent from '@mui/material/DialogContent'
import Typography from '@mui/material/Typography'
import Error from '@mui/icons-material/Error'
import { styles as globalStyles } from '../../styles/materialStyles'
import { Zoom } from '@mui/material'
import { useTheme } from '@mui/material/styles'
function DialogContentError(props) {
  const theme = useTheme()
  const { error: errorSx } = globalStyles(theme)
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
          <Error sx={errorSx} />
          &nbsp;
          <span style={errorSx}>{error}</span>
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
