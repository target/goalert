# Form Components

Form components handle passing validation and value information to separate 3 main concerns:

1. Errors and value data/state -- `FormContainer`
2. Individual field layout and validation -- `FormField`
3. Overall form validation & submitting -- `Form`

### FormContainer

The `FormContainer` component handles a single form segment. It takes a `value` prop as an object, where the keys
map to descendant `FormField` components. Rather than value, error, onChange, etc.. for each FormField, a `FormContainer` will take an `errors` array and a value and matching onChange handler that deal with a single object.

- If a `FormContainer` is `disabled` all fields within it will be as well.
- A `FormContainer` will run validation if an outer `Form` component fires it's `onSubmit` handler. Validation passes if all nested `FormField` components are valid.

### FormField

The `FormField` component will receive `value` and `error` data from a `FormContainer`. It also registers a `validate`
function, if provided. The field name (used for error checking and the container's value prop) defaults to `name` but can be overridden with the `fieldName` prop.

### Form

The `Form` component's job is to simply report when it has been submitted and indicate whether or not all `FormContainer` components pass validation.

## Basic Usage

The following example will only call `doMutation` if the user enters 1 or more digits
into the `Foo` text field.

If the form is submitted with no value, the user will be directed to enter a value.
If the form is submitted with non-digits, an error will be displayed in the form field "Only numbers allowed.".

```js
<Form
  onSubmit={(event, isValid) => {
    e.preventDefault()
    // isValid will be true if all FormContainers report no validation errors
    if (isValid) doMutation()
  }}
>
  <FormContainer value={{ foo: 'bar' }}>
    <FormField
      component={TextField}
      name='foo'
      label='Foo'
      // required indicates the field must not be empty
      required
      // validate allows custom validation & error messages
      validate={value =>
        /^\d+$/.test(value) ? null : new Error('Only numbers allowed.')
      }
    />
  </FormContainer>
  // ... Multiple FormContainers can be placed in the same form (e.g. SetupWizard)
</Form>
```
