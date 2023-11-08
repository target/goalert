import React from 'react'
import { useQuery, gql } from 'urql'
import p from 'prop-types'
import MenuItem from '@mui/material/MenuItem'
import TextField from '@mui/material/TextField'
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
  const [{ data }] = useQuery({
    query,
    requestPolicy: 'network-only',
    variables: {
      id: userID,
    },
  })

  const cms = data?.user ? data?.user?.contactMethods : []

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
