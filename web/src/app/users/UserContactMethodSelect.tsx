import React, { ReactNode } from 'react'
import { useQuery, gql } from 'urql'
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

type CMExtraItems = {
  label?: ReactNode
  value: string
}

interface UserContactMethodSelectProps {
  userID: string
  extraItems: CMExtraItems[]
}

export default function UserContactMethodSelect({
  userID,
  extraItems = [] as CMExtraItems[],
  ...rest
}: UserContactMethodSelectProps): ReactNode {
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
