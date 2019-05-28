import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import { withStyles } from '@material-ui/core/styles'
import Select from 'react-select'
import { components, styles } from './MaterialSelectComponents'

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

@withStyles(styles, { withTheme: true })
export default class MaterialSelect extends Component {
  static propTypes = {
    multiple: p.bool, // allow selecting multiple values
    required: p.bool,
    onChange: p.func.isRequired,
    onInputChange: p.func,
    options: p.arrayOf(valueShape).isRequired,
    placeholder: p.string,
    value: valueCheck,
  }

  static defaultProps = {
    options: [],
  }

  render() {
    const {
      multiple,
      noClear,
      theme,
      value,
      onChange,
      classes,
      disabled,
      required,

      label,
      name,
      placeholder,
      InputLabelProps,

      ...props
    } = this.props

    const selectStyles = {
      input: base => ({
        ...base,
        color: theme.palette.text.primary,
      }),
    }

    let textFieldProps = {
      required,
      label,
      placeholder,
      InputLabelProps,
      value: value ? (multiple ? value.join(',') : value.value) : '',
    }

    return (
      <div
        data-cy='material-select'
        data-cy-ready={!props.isLoading}
        className={classes.root}
      >
        <Select
          menuPortalTarget={document.body}
          menuPlacement='auto'
          styles={{
            ...selectStyles,
            menuPortal: base => ({ ...base, zIndex: 9999 }),
          }}
          name={name}
          classes={classes}
          isClearable={!required}
          isDisabled={disabled}
          isMulti={multiple}
          value={value}
          components={components}
          onChange={onChange}
          textFieldProps={textFieldProps}
          placeholder=''
          {...props}
        />
      </div>
    )
  }
}
