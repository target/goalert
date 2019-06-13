import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import IconButton from '@material-ui/core/IconButton'
import Input from '@material-ui/core/Input'
import InputAdornment from '@material-ui/core/InputAdornment'
import Switch from '@material-ui/core/Switch'
import Visibility from '@material-ui/icons/Visibility'
import VisibilityOff from '@material-ui/icons/VisibilityOff'

export const StringListInput = props => {
  const value = props.value ? props.value.split('\n').concat('') : ['']
  return (
    <Grid container spacing={1}>
      {value.map((val, idx) => (
        <Grid key={idx} item xs={12}>
          <StringInput
            value={val}
            name={val ? props.name + '-' + idx : props.name + '-new-item'}
            onChange={newVal =>
              props.onChange(
                value
                  .slice(0, idx)
                  .concat(newVal, ...value.slice(idx + 1))
                  .filter(v => v)
                  .join('\n'),
              )
            }
            autoComplete='new-password'
            password={props.password}
          />
        </Grid>
      ))}
    </Grid>
  )
}

export class StringInput extends React.PureComponent {
  static propTypes = {
    password: p.bool,
  }

  state = {
    showPassword: false,
  }

  render() {
    const { onChange, password, ...rest } = this.props

    return (
      <Input
        fullWidth
        autoComplete='new-password' // chrome keeps autofilling them, this stops it
        type={password && !this.state.showPassword ? 'password' : 'text'}
        onChange={e => onChange(e.target.value)}
        endAdornment={this.renderPasswordAdornment()}
        {...rest}
      />
    )
  }

  renderPasswordAdornment() {
    if (!this.props.password) return null

    return (
      <InputAdornment position='end'>
        <IconButton
          aria-label='Toggle password visibility'
          onClick={() =>
            this.setState({ showPassword: !this.state.showPassword })
          }
        >
          {this.state.showPassword ? <Visibility /> : <VisibilityOff />}
        </IconButton>
      </InputAdornment>
    )
  }
}

export const IntegerInput = props => (
  <Input
    {...props}
    type='number'
    fullWidth
    onChange={e => props.onChange(e.target.value)}
  />
)

export const BoolInput = props => (
  <Switch
    {...props}
    checked={props.value === 'true'}
    onChange={e => props.onChange(e.target.checked ? 'true' : 'false')}
  />
)
