import React, {
  useEffect,
  useState,
  ReactNode,
  ReactElement,
  SyntheticEvent,
} from 'react'
import {
  TextField,
  MenuItem,
  ListItemIcon,
  Typography,
  Paper,
  Chip,
  InputProps,
  Alert,
  Autocomplete,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'

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
  clearIndicator: {
    display: 'none',
  },
  padding0: {
    padding: 0,
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
  noOptionsError?: Error
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
    noOptionsError,
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

  const customCSS: Record<string, string> = {
    option: classes.padding0,
    clearIndicator: classes.clearIndicator,
  }

  if (noOptionsError) {
    customCSS.noOptions = classes.padding0
  }

  function isSelected(val: string): boolean {
    if (!value) return false

    if (Array.isArray(value)) {
      let isIncluded = false
      value.forEach((value) => {
        if (val === value.value) {
          isIncluded = true
        }
      })
      return isIncluded
    }

    return val === value.value
  }

  return (
    <Autocomplete
      data-cy='material-select'
      data-cy-ready={!isLoading}
      classes={customCSS}
      {...multi}
      value={value}
      filterSelectedOptions
      inputValue={inputValue}
      disableClearable={required}
      disabled={disabled}
      isOptionEqualToValue={(opt, val) => opt.value === val.value}
      noOptionsText={
        noOptionsError ? (
          <Alert severity='error'>{noOptionsError.message}</Alert>
        ) : (
          noOptionsText
        )
      }
      onChange={(
        event: SyntheticEvent<Element, Event>,
        selected: SelectOption | SelectOption[] | null,
      ) => {
        if (selected) {
          if (Array.isArray(selected)) {
            setInputValue('') // clear input so user can keep typing to select another item
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
      options={_options}
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
      renderOption={(props, { label, icon, value }) => (
        <MenuItem
          {...props}
          component='span'
          className={classes.menuItem}
          selected={isSelected(value)}
          data-cy='search-select-item'
        >
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
            {...getTagProps({ index })}
            key={index.toString()}
            data-cy='multi-value'
            label={option.label}
          />
        ))
      }
    />
  )
}
