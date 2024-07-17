import React from 'react'
import Button, { ButtonProps } from '@mui/material/Button'
import Tooltip from '@mui/material/Tooltip'
import CircularProgress from '@mui/material/CircularProgress'

interface LoadingButtonProps extends ButtonProps {
  attemptCount?: number
  buttonText?: string
  loading?: boolean
  noSubmit?: boolean
  tooltip?: string
  style?: React.CSSProperties
}

const LoadingButton = (props: LoadingButtonProps): JSX.Element => {
  const {
    attemptCount,
    buttonText,
    color,
    disabled,
    loading,
    noSubmit,
    onClick,
    tooltip,
    style,
    ...rest
  } = props

  const button = (
    <Button
      data-cy='loading-button'
      variant='contained'
      {...rest}
      color={color || 'primary'}
      onClick={onClick}
      disabled={loading || disabled}
      type={noSubmit ? 'button' : 'submit'}
    >
      {!attemptCount ? props.children || buttonText || 'Confirm' : 'Retry'}
    </Button>
  )

  return (
    <div style={{ position: 'relative', ...style }}>
      {tooltip ? <Tooltip title={tooltip}>{button}</Tooltip> : button}
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

export default LoadingButton
