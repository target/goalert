import React from 'react'
import p from 'prop-types'
import Button from '@material-ui/core/Button'
import CircularProgress from '@material-ui/core/CircularProgress'

const LoadingButton = (props) => {
  const {
    attemptCount,
    buttonText,
    color,
    disabled,
    loading,
    noSubmit,
    onClick,
    style,
    ...rest
  } = props

  return (
    <div style={{ position: 'relative', ...style }}>
      <Button
        {...rest}
        data-cy='loading-button'
        variant='contained'
        color={color || 'primary'}
        onClick={onClick}
        disabled={loading || disabled}
        type={noSubmit ? 'button' : 'submit'}
      >
        {!attemptCount ? buttonText || 'Confirm' : 'Retry'}
      </Button>
      {loading && (
        <CircularProgress
          color={color || 'primary'}
          size={24}
          style={{
            position: 'absolute',
            top: '50%',
            left: '50%',
            marginTop: -12,
            marginLeft: -12,
          }}
        />
      )}
    </div>
  )
}

LoadingButton.propTypes = {
  attemptCount: p.number,
  buttonText: p.string,
  color: p.string,
  disabled: p.bool,
  loading: p.bool,
  noSubmit: p.bool,
  onClick: p.func,
}

export default LoadingButton
