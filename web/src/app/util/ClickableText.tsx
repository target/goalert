import { makeStyles } from '@material-ui/core'
import React, { ButtonHTMLAttributes, MouseEventHandler } from 'react'

const crimson = '#cd1831'

const useStyles = makeStyles({
  button: {
    color: crimson,
    all: 'unset',
    backgroundColor: 'transparent',
    border: 'none',
    cursor: 'pointer',
    textDecoration: 'none',
    display: 'inline',
    margin: 0,
    padding: 0,
    '&:hover': {
      textDecoration: 'underline',
    },
    '&:focus': {
      textDecoration: 'underline',
    },
  },
})

interface ClickableTextProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  text: string
  onClick: MouseEventHandler<HTMLButtonElement>
}

// ClickableText
// since anchor elements have become entrenched with navigation by convention
// it is recommended that clickable text be implemented using a button element
// see https://github.com/jsx-eslint/eslint-plugin-jsx-a11y/blob/master/docs/rules/anchor-is-valid.md
function ClickableText({
  text,
  type = 'button',
  ...props
}: ClickableTextProps): JSX.Element {
  const classes = useStyles()
  return (
    <button className={classes.button} type={type} {...props}>
      {text}
    </button>
  )
}

export default ClickableText
