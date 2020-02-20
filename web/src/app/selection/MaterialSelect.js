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

export const styles = () => ({
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
  option: {
    padding: 0,
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

  // let value = propsValue
  // if (!value) {
  //   if (multiple) value = []
  //   else value = { label: '', value: '' }
  // }
  let value
  if (multiple && !propsValue) value = []
  else if (!multiple && !propsValue) value = { label: '', value: '' }
  else if (multiple && propsValue) value = propsValue
  else if (!multiple && propsValue) value = propsValue

  const [inputValue, setInputValue] = useState(multiple ? '' : value.label)

  return (
    <div data-cy='material-select' data-cy-ready={!isLoading}>
      <Autocomplete
        classes={{ option: classes.option }}
        value={multiple ? value : value.value}
        inputValue={inputValue}
        disableClearable={required}
        disabled={disabled}
        multiple={multiple}
        filterSelectedOptions
        onChange={(event, valueObj) => {
          if (valueObj === null) {
            onChange(null)
          } else {
            onChange(valueObj)
            setInputValue(multiple ? '' : valueObj.label)
          }
        }}
        onInputChange={(event, inputVal, reason) => {
          if (reason === 'clear' && !multiple) {
            setInputValue('')
          }
        }}
        onBlur={() => setInputValue(multiple ? '' : value.label)}
        loading={isLoading}
        getOptionLabel={option => option.label || ''}
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
              data-cy='search-select'
              fullWidth
              label={label}
              onChange={({ target }) => {
                const newInputVal = target.value
                setInputValue(newInputVal)
                if (onInputChange) onInputChange(newInputVal)
              }}
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
