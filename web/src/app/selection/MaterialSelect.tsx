import React, {
  useEffect,
  useState,
  ReactNode,
  ReactElement,
  ChangeEvent,
} from 'react'
import {
  TextField,
  makeStyles,
  MenuItem,
  ListItemIcon,
  Typography,
  Paper,
  Chip,
  InputProps,
} from '@material-ui/core'
import { Autocomplete } from '@material-ui/lab'

const useStyles = makeStyles({
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
  clearIndicator: {
    display: 'none',
  },
})

interface AutocompleteInputProps extends InputProps {
  'data-cy': string
}

interface SelectOption {
  icon?: ReactElement
  isCreate?: boolean
  label: string
  value: string
}

interface CommonSelectProps {
  disabled?: boolean
  error?: boolean
  isLoading?: boolean
  label?: string
  noOptionsText?: ReactNode
  name?: string
  required?: boolean
  onInputChange?: (value: string) => void
  options: SelectOption[]
  placeholder?: string
}

interface SingleSelectProps {
  multiple: false
  value: SelectOption
  onChange: (value: SelectOption | null) => void
}

interface MultiSelectProps {
  multiple: true
  value: SelectOption[]
  onChange: (value: SelectOption[]) => void
}

export default function MaterialSelect(
  props: CommonSelectProps & (MultiSelectProps | SingleSelectProps),
): JSX.Element {
  const classes = useStyles()
  const {
    disabled,
    error,
    isLoading,
    label,
    multiple,
    name,
    noOptionsText,
    onChange,
    onInputChange = () => {},
    options: _options,
    placeholder,
    required,
    value,
  } = props

  // getInputLabel will return the label of the current value.
  //
  // If in multi-select mode an empty string is always returned as selected values
  // are never preserved in the input field (they are chips instead).
  const getInputLabel = (): string =>
    multiple || Array.isArray(value) ? '' : value?.label || ''

  const [focus, setFocus] = useState(false)
  const [inputValue, _setInputValue] = useState(getInputLabel())

  const setInputValue = (input: string): void => {
    _setInputValue(input)
    onInputChange(input)
  }

  const multi = multiple ? { multiple: true, filterSelectedOptions: true } : {}

  useEffect(() => {
    if (!focus) setInputValue(getInputLabel())
    if (multiple) return
    if (!value) setInputValue('')
  }, [value, multiple, focus])

  // merge selected values with options to avoid annoying mui warnings
  // https://github.com/mui-org/material-ui/issues/18514
  let options = _options
  if (value && Array.isArray(value)) {
    options = [...options, ...value]
  } else if (
    value &&
    !Array.isArray(value) &&
    options.every((opt) => opt.value !== value.value)
  ) {
    options = [value, ...options]
  }

  return (
    <Autocomplete
      data-cy='material-select'
      data-cy-ready={!isLoading}
      classes={{
        option: classes.option,
        clearIndicator: classes.clearIndicator,
      }}
      {...multi}
      value={value}
      inputValue={inputValue}
      disableClearable={required}
      disabled={disabled}
      getOptionSelected={(opt, val) => opt.value === val.value}
      noOptionsText={noOptionsText}
      onChange={(
        event: ChangeEvent<Record<string, unknown>>,
        selected: SelectOption | SelectOption[] | null,
      ) => {
        if (selected) {
          if (Array.isArray(selected)) {
            setInputValue('')
          } else {
            setInputValue(selected.isCreate ? selected.value : selected.label)
          }
        }

        // NOTE typeof selected switches based on multiple; ts can't infer this
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        onChange(selected as any)
      }}
      onInputChange={(event, inputVal, reason) => {
        if (reason === 'clear' && !multiple) {
          setInputValue('')
        }
      }}
      onFocus={() => setFocus(true)}
      onBlur={() => setFocus(false)}
      loading={isLoading}
      getOptionLabel={(option) => option?.label ?? ''}
      options={options}
      renderInput={(params) => {
        return (
          <TextField
            {...params}
            inputProps={{
              ...params.inputProps,
              name,
            }}
            InputProps={
              {
                ...params.InputProps,
                'data-cy': 'search-select-input',
              } as AutocompleteInputProps
            }
            data-cy='search-select'
            fullWidth
            label={label}
            placeholder={placeholder}
            onChange={({ target }) => {
              const newInputVal: string = target.value
              setInputValue(newInputVal)
            }}
            error={error}
          />
        )
      }}
      renderOption={({ label, icon }) => (
        <MenuItem component='span' className={classes.menuItem}>
          <Typography noWrap>{label}</Typography>
          {icon && (
            <ListItemIcon className={classes.listItemIcon}>{icon}</ListItemIcon>
          )}
        </MenuItem>
      )}
      PaperComponent={(params) => (
        <Paper data-cy='select-dropdown' {...params} />
      )}
      renderTags={(value, getTagProps) =>
        value.map((option, index) => (
          <Chip
            key={index.toString()}
            data-cy='multi-value'
            label={option.label}
            {...getTagProps({ index })}
          />
        ))
      }
    />
  )
}
