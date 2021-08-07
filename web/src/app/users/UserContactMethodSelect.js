import React from 'react'
import { gql } from '@apollo/client'
import p from 'prop-types'
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

export default function UserContactMethodSelect({
  userID,
  extraItems,
  ...rest
}) {
  function renderControl(cms) {
    return (
      <TextField select {...rest}>
        {sortContactMethods(cms)
          .map((cm) => (
            <MenuItem key={cm.id} value={cm.id}>
              {cm.name} ({cm.type})
            </MenuItem>
          ))
          .concat(
            extraItems.map((item) => (
              <MenuItem key={item.value} value={item.value}>
                {item.label || item.value}
              </MenuItem>
            )),
          )}
      </TextField>
    )
  }

  return (
    <Query
      query={query}
      variables={{ id: userID }}
      render={({ data }) => renderControl(data.user.contactMethods)}
    />
  )
}

UserContactMethodSelect.propTypes = {
  userID: p.string.isRequired,

  extraItems: p.arrayOf(
    p.shape({
      label: p.node,
      value: p.string.isRequired,
    }),
  ),
}

UserContactMethodSelect.defaultProps = {
  extraItems: [],
}
