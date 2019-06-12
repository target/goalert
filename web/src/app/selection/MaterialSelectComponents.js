import React from 'react'
import classNames from 'classnames'
import Typography from '@material-ui/core/Typography'
import TextField from '@material-ui/core/TextField'
import Paper from '@material-ui/core/Paper'
import Chip from '@material-ui/core/Chip'
import MenuItem from '@material-ui/core/MenuItem'
import { emphasize } from '@material-ui/core/styles/colorManipulator'
import ListItemIcon from '@material-ui/core/ListItemIcon'

export const styles = theme => ({
  root: {
    flexGrow: 1,
  },
  input: {
    display: 'flex',
    padding: 0,
    height: 'fit-content',
  },
  valueContainer: {
    display: 'flex',
    flexWrap: 'wrap',
    flex: 1,
    alignItems: 'center',
  },
  chip: {
    margin: `${theme.spacing(1) / 2}px ${theme.spacing(1) / 4}px`,
  },
  chipFocused: {
    backgroundColor: emphasize(
      theme.palette.type === 'light'
        ? theme.palette.grey[300]
        : theme.palette.grey[700],
      0.08,
    ),
  },
  noOptionsMessage: {
    padding: `${theme.spacing(1)}px ${theme.spacing(2)}px`,
  },
  singleValue: {
    fontSize: 16,
  },
  placeholder: {
    position: 'absolute',
    left: 2,
    fontSize: 16,
  },
  paper: {
    position: 'absolute',
    zIndex: 1,
    marginTop: theme.spacing(1),
    left: 0,
    right: 0,
  },
})

export const NoOptionsMessage = props => (
  <Typography
    color='textSecondary'
    className={props.selectProps.classes.noOptionsMessage}
    {...props.innerProps}
  >
    {props.children}
  </Typography>
)

export const inputComponent = ({ inputRef, ...props }) => (
  <div ref={inputRef} {...props} />
)

export const Control = props => (
  <TextField
    data-cy='search-select'
    disabled={props.isDisabled}
    error={props.selectProps.error}
    fullWidth
    InputProps={{
      inputComponent,
      inputProps: {
        className: props.selectProps.classes.input,
        children: props.children,
        inputRef: props.innerRef,
        name: props.selectProps.name,
        'data-cy': 'search-select-input',
        ...props.innerProps,
      },
    }}
    {...props.selectProps.textFieldProps}
  />
)

export const Option = props => (
  <MenuItem
    buttonRef={props.innerRef}
    selected={props.isFocused}
    component='div'
    style={{
      fontWeight: props.isSelected ? 500 : 400,
    }}
    {...props.innerProps}
  >
    {props.children}
    {props.data.icon && (
      <ListItemIcon style={{ position: 'absolute', right: 0 }}>
        {props.data.icon}
      </ListItemIcon>
    )}
  </MenuItem>
)

export const Placeholder = props => (
  <Typography
    color={props.selectProps.error ? 'error' : 'textSecondary'}
    className={props.selectProps.classes.placeholder}
    {...props.innerProps}
  >
    {props.children}
  </Typography>
)

export const ValueContainer = props => (
  <div className={props.selectProps.classes.valueContainer}>
    {props.children}
  </div>
)

export const SingleValue = props => (
  <Typography
    color={props.isDisabled ? 'textSecondary' : 'initial'}
    className={props.selectProps.classes.singleValue}
    {...props.innerProps}
  >
    {props.children}
  </Typography>
)

export const MultiValue = props => (
  <Chip
    data-cy='multi-value'
    tabIndex={-1}
    label={props.children}
    className={classNames(props.selectProps.classes.chip, {
      [props.selectProps.classes.chipFocused]: props.isFocused,
    })}
    onDelete={event => {
      props.removeProps.onClick()
      props.removeProps.onMouseDown(event)
    }}
  />
)

export const Menu = props => (
  <Paper
    className={props.selectProps.classes.paper}
    data-cy='select-dropdown'
    square
    {...props.innerProps}
  >
    {props.children}
  </Paper>
)

export const components = {
  Option,
  Control,
  NoOptionsMessage,
  Placeholder,
  SingleValue,
  MultiValue,
  ValueContainer,
  Menu,
}
