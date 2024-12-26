import React from 'react'
import makeStyles from '@mui/styles/makeStyles'
import Button, { ButtonProps } from '@mui/material/Button'

const useStyles = makeStyles({
  button: {
    fontWeight: 'normal', // disable MUI bold
    padding: '1px',
    textTransform: 'none', // disable MUI uppercase
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
}: ButtonProps): React.JSX.Element {
  const classes = useStyles()
  return (
    <Button
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
