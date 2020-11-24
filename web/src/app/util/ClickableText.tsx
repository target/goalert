import React from 'react'
import makeStyles from '@material-ui/core/styles/makeStyles'
import Button, { ButtonProps } from '@material-ui/core/Button'

const useStyles = makeStyles({
  button: {
    fontWeight: 'normal', // disable MUI bold
    textTransform: 'none', // disable MUI uppercase
    margin: 0,
    padding: 0,
    '&:hover': {
      textDecoration: 'underline',
      background: 'transparent',
    },
    '&:focus': {
      textDecoration: 'underline',
    },
  },
})

// ClickableText utilizes MUI's outlined button while adding some default  stylistic changes
// since anchor elements have become entrenched with navigation by convention
//
// it's recommended that clickable text is implemented using a button element
// see https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/master/docs/rules/anchor-is-valid.md
export default function ClickableText({
  children,
  color = 'secondary',
  size = 'small',
  type = 'button',
  ...props
}: ButtonProps): JSX.Element {
  const classes = useStyles()
  return (
    <Button
      disableRipple
      disableFocusRipple
      className={classes.button}
      color={color}
      size={size}
      type={type}
      {...props}
    >
      {children}
    </Button>
  )
}
