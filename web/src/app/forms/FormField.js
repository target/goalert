import React from 'react'
import p from 'prop-types'
import MountWatcher from '../util/MountWatcher'
import FormControl from '@material-ui/core/FormControl'
import FormHelperText from '@material-ui/core/FormHelperText'
import FormLabel from '@material-ui/core/FormLabel'
import { get, isEmpty, startCase } from 'lodash'
import shrinkWorkaround from '../util/shrinkWorkaround'
import AppLink from '../util/AppLink'
import { FormContainerContext } from './context'

export class FormField extends React.PureComponent {
  static propTypes = {
    // one of component or render must be provided
    component: p.any,
    render: p.func,

    // mapValue can be used to map a value before it's passed to the form component
    mapValue: p.func,

    // mapOnChangeValue can be used to map a changed value from the component, before it's
    // passed to the parent form's state.
    mapOnChangeValue: p.func,

    // Adjusts props for usage with a Checkbox component.
    checkbox: p.bool,

    // fieldName specifies the field used for
    // checking errors, change handlers, and value.
    //
    // If unset, it defaults to `name`.
    name: p.string.isRequired,
    fieldName: p.string,

    // min and max values specify the range to clamp a int value
    // expects an ISO timestamp, if string
    min: p.oneOfType([p.number, p.string]),
    max: p.oneOfType([p.number, p.string]),

    // used if name is set,
    // but the error name is different from graphql responses
    errorName: p.string,

    // label above form component
    label: p.node,
    formLabel: p.bool, // use formLabel instead of label if true

    // required indicates the field may not be left blank.
    required: p.bool,

    // validate can be used to provide client-side validation of a
    // field.
    validate: p.func,

    // a hint for the user on a form field. errors take priority
    hint: p.string,

    // disable the form helper text for errors.
    noError: p.bool,
  }

  static defaultProps = {
    validate: () => {},
    mapValue: (value) => value,
    mapOnChangeValue: (value) => value,
  }

  validate = (value) => {
    if (
      this.props.required &&
      !['boolean', 'number'].includes(typeof value) &&
      isEmpty(value)
    ) {
      return new Error('Required field.')
    }

    return this.props.validate(value)
  }

  render() {
    return (
      <FormContainerContext.Consumer>
        {this.renderComponent}
      </FormContainerContext.Consumer>
    )
  }

  renderComponent = ({
    errors,
    value,
    onChange,
    addField,
    disabled: containerDisabled,
    optionalLabels,
    ...otherFormProps
  }) => {
    const {
      errorName,
      name,
      noError,
      component: Component,
      render,
      fieldName: _fieldName,
      formLabel,
      required,
      validate,
      disabled: fieldDisabled,
      hint,
      label: _label,
      InputLabelProps: _inputProps,
      mapValue,
      mapOnChangeValue,
      min,
      max,
      checkbox,
      ...otherFieldProps
    } = this.props

    const baseLabel = typeof _label === 'string' ? _label : startCase(name)
    const label =
      !required && optionalLabels ? baseLabel + ' (optional)' : baseLabel

    const fieldName = _fieldName || name
    const props = {
      ...otherFormProps,
      ...otherFieldProps,
      name,
      required,
      disabled: containerDisabled || fieldDisabled,
      error: errors.find((err) => err.field === (errorName || fieldName)),
      hint,
      value: mapValue(get(value, fieldName)),
      min,
      max,
    }

    const InputLabelProps = {
      required: required && !optionalLabels,
      ...shrinkWorkaround(props.value),
      ..._inputProps,
    }

    let getValueOf = (e) => (e && e.target ? e.target.value : e)
    if (checkbox) {
      props.checked = props.value
      props.value = props.value ? 'true' : 'false'
      getValueOf = (e) => e.target.checked
    } else if (otherFieldProps.type === 'number') {
      props.label = label
      props.value = props.value.toString()
      props.InputLabelProps = InputLabelProps
      getValueOf = (e) => parseInt(e.target.value, 10)
    } else {
      props.label = label
      props.InputLabelProps = InputLabelProps
    }

    props.onChange = (value) => {
      let newValue = getValueOf(value)
      if (props.type === 'number' && typeof props.min === 'number')
        newValue = Math.max(props.min, newValue)
      if (props.type === 'number' && typeof props.max === 'number')
        newValue = Math.min(props.max, newValue)
      onChange(fieldName, mapOnChangeValue(newValue))
    }

    return (
      <MountWatcher
        onMount={() => {
          this._unregister = addField(fieldName, this.validate)
        }}
        onUnmount={() => {
          this._unregister()
        }}
      >
        {this.renderContent(props)}
      </MountWatcher>
    )
  }

  renderContent(props) {
    const {
      checkbox,
      component,
      formLabel,
      label,
      noError,
      render,
    } = this.props

    if (render) return render(props)
    const Component = component

    return (
      <FormControl fullWidth={props.fullWidth} error={Boolean(props.error)}>
        {formLabel && (
          <FormLabel style={{ paddingBottom: '0.5em' }}>{label}</FormLabel>
        )}
        <Component
          {...props}
          error={checkbox ? undefined : Boolean(props.error)}
          label={this.props.formLabel ? null : props.label}
        />
        {!noError && this.renderFormHelperText(props.error, props.hint)}
      </FormControl>
    )
  }

  renderFormHelperText(error, hint) {
    if (error?.helpLink) {
      return (
        <FormHelperText>
          <AppLink to={error.helpLink} newTab data-cy='error-help-link'>
            {error.message.replace(/^./, (str) => str.toUpperCase())}
          </AppLink>
        </FormHelperText>
      )
    }

    if (error?.message) {
      return (
        <FormHelperText>
          {error.message.replace(/^./, (str) => str.toUpperCase())}
        </FormHelperText>
      )
    }

    if (hint) {
      return <FormHelperText>{hint}</FormHelperText>
    }

    return null
  }
}
