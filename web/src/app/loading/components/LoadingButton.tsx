import React, { ReactElement } from 'react'
import Button, { ButtonPropsColorOverrides } from '@mui/material/Button'
import CircularProgress from '@mui/material/CircularProgress'
import { OverridableStringUnion } from '@mui/types'

interface LoadingButtonProps {
  attemptCount?: number
  buttonText?: string
  color?: OverridableStringUnion<
    | 'inherit'
    | 'primary'
    | 'secondary'
    | 'success'
    | 'error'
    | 'info'
    | 'warning',
    ButtonPropsColorOverrides
  >
  disabled?: boolean
  loading?: boolean
  noSubmit?: boolean
  onClick?: () => void
  style?: React.CSSProperties
}

const LoadingButton = (props: LoadingButtonProps): ReactElement => {
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

export default LoadingButton
