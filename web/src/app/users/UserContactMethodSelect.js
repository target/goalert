import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import MenuItem from '@material-ui/core/MenuItem'
import TextField from '@material-ui/core/TextField'
import Query from '../util/Query'
import { sortContactMethods } from './util'

const query = gql`
  query userCMSelect($id: ID!) {
    user(id: $id) {
      id
      contactMethods {
        id
        name
        type
        value
      }
    }
  }
`

export default class UserContactMethodSelect extends React.PureComponent {
  static propTypes = {
    userID: p.string.isRequired,

    extraItems: p.arrayOf(
      p.shape({
        label: p.node,
        value: p.string.isRequired,
      }),
    ),
  }

  static defaultProps = {
    extraItems: [],
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.userID }}
        render={({ data }) => this.renderControl(data.user.contactMethods)}
      />
    )
  }

  renderControl(cms) {
    const { userID, extraItems, ...rest } = this.props

    return (
      <TextField select {...rest}>
        {sortContactMethods(cms)
          .map(cm => (
            <MenuItem key={cm.id} value={cm.id}>
              {cm.name} ({cm.type})
            </MenuItem>
          ))
          .concat(
            extraItems.map(item => (
              <MenuItem key={item.value} value={item.value}>
                {item.label || item.value}
              </MenuItem>
            )),
          )}
      </TextField>
    )
  }
}
