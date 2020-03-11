import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import { withStyles } from '@material-ui/core/styles'
import Select from 'react-select'
import { components, styles } from './MaterialSelectComponents'
import shrinkWorkaround from '../util/shrinkWorkaround'
import _ from 'lodash-es'

const valueShape = p.shape({
  label: p.string.isRequired,
  value: p.any.isRequired,
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
  constructor(props) {
    super(props)
    this.clearButtonRef = React.createRef()
    this.state = {
      isCleared: false,
    }
    this.buttonClicked = false
  }

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
      theme,
      value,
      onChange,
      onInputChange,
      classes,
      disabled,
      required,

      label,
      name,
      placeholder,
      InputLabelProps: _InputLabelProps,

      ...props
    } = this.props

    const InputLabelProps = {
      ...shrinkWorkaround(this.props.value),
      ..._InputLabelProps,
    }

    const selectStyles = {
      input: base => ({
        ...base,
        color: theme.palette.text.primary,
      }),
    }

    const textFieldProps = {
      required,
      label,
      placeholder,
      InputLabelProps,
      value: value ? (multiple ? value.join(',') : value.value) : '',
    }
    const { isCleared } = this.state

    const isDescendent = (parent, child) => {
      if (!child) return false
      if (child.isSameNode(parent)) return true
      return isDescendent(parent, child.parentNode)
    }

    return (
      // going to look into replacing component with material-ui autocomplete component
      // eslint-disable-next-line jsx-a11y/click-events-have-key-events,jsx-a11y/no-static-element-interactions
      <div
        data-cy='material-select'
        data-cy-ready={!props.isLoading}
        className={classes.root}
        onMouseDownCapture={e => {
          if (isDescendent(this.clearButtonRef.current, e.target))
            this.buttonClicked = true
        }}
        onClickCapture={() => {
          this.buttonClicked = false
        }}
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
          isClearable
          isDisabled={disabled}
          isMulti={multiple}
          value={isCleared && required ? { label: '', value: '' } : value}
          clearButtonRef={this.clearButtonRef}
          components={components}
          onBlur={() => {
            if (!this.buttonClicked) this.setState({ isCleared: false })
          }}
          onChange={val => {
            if (required && _.isEmpty(val) && !multiple) {
              this.setState({ isCleared: true })
              return
            }
            if (required && val !== null && !multiple)
              this.setState({ isCleared: false })
            onChange(val)
          }}
          onInputChange={val => {
            this.buttonClicked = false
            if (onInputChange) onInputChange(val)
          }}
          textFieldProps={textFieldProps}
          placeholder=''
          {...props}
        />
      </div>
    )
  }
}
