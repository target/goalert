import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import {
  TextField,
  makeStyles,
  MenuItem,
  ListItemIcon,
  Typography,
} from '@material-ui/core'
import { Autocomplete } from '@material-ui/lab'
import { emphasize } from '@material-ui/core/styles/colorManipulator'

export const styles = theme => ({
  root: {
    flexGrow: 1,
  },
  option: {
    padding: 0,
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
    wordBreak: 'break-word',
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
  listItemIcon: {
    minWidth: 0,
  },
  menuItem: {
    display: 'flex',
    flex: 1,
    justifyContent: 'space-between',
    wordBreak: 'break-word',
    whiteSpace: 'pre-wrap',
  },
  message: {
    float: 'left',
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

const useStyles = makeStyles(styles)

export default function MaterialSelect(props) {
  const classes = useStyles()
  const {
    disabled,
    // fullWidth,
    // hint,
    isLoading,
    label,
    multiple,
    name,
    // noOptionsMessage,
    onChange,
    onInputChange,
    options,
    required,
    // theme,
    value: propsValue,

    // classes
    // placeholder: string
    // error: boolean
  } = props

  let value = { value: '', label: '' }
  if (propsValue !== null) {
    value = propsValue
  }

  const [inputValue, setInputValue] = useState(value.label)

  return (
    <div data-cy='material-select' data-cy-ready={!isLoading}>
      <Autocomplete
        classes={{ option: classes.option }}
        value={value.value}
        inputValue={inputValue}
        disableClearable={required}
        disabled={disabled}
        multiple={multiple}
        onChange={(event, valueObj) => {
          if (valueObj === null) {
            onChange(null)
          } else {
            onChange(valueObj)
            setInputValue(valueObj.label)
          }
        }}
        onBlur={() => {
          if (required) setInputValue(value.label)
        }}
        onInputChange={(event, value, reason) => {
          if (reason === 'clear') {
            setInputValue('')
          }
        }}
        loading={isLoading}
        getOptionLabel={option => option.label || 'Loading...'}
        options={options}
        renderInput={params => {
          return (
            <TextField
              {...params}
              inputProps={{
                ...params.inputProps,
                name,
                'data-cy': 'search-select-input',
              }}
              onChange={event => {
                setInputValue(event.target.value)
                if (onInputChange) onInputChange(event.target.value)
              }}
              data-cy='search-select'
              fullWidth
              label={label}
            />
          )
        }}
        renderOption={({ label, value, icon }) => (
          <MenuItem component='span' className={classes.menuItem}>
            <Typography noWrap>{label}</Typography>
            {icon && (
              <ListItemIcon className={classes.listItemIcon}>
                {icon}
              </ListItemIcon>
            )}
          </MenuItem>
        )}
      />
    </div>
  )
}

const valueShape = p.shape({
  label: p.string.isRequired,
  value: p.string.isRequired,
  icon: p.element,
})

// valueCheck ensures the type is `arrayOf(p.string)` if `multiple` is set
// and `p.string` otherwise.
function valueCheck(props, ...args) {
  if (props.multiple) return p.arrayOf(valueShape).isRequired(props, ...args)
  return valueShape(props, ...args)
}

MaterialSelect.propTypes = {
  multiple: p.bool, // allow selecting multiple values
  required: p.bool,
  onChange: p.func.isRequired,
  onInputChange: p.func,
  options: p.arrayOf(valueShape).isRequired,
  placeholder: p.string,
  value: valueCheck,
}

MaterialSelect.defaultProps = {
  options: [],
}
