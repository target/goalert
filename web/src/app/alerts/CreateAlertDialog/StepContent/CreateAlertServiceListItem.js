import React from 'react'
import p from 'prop-types'
import { useQuery } from 'react-apollo'
import {
  ListItem,
  ListItemText,
  Typography,
  makeStyles,
} from '@material-ui/core'
import OpenInNewIcon from '@material-ui/icons/OpenInNew'

import gql from 'graphql-tag'
import AppLink from '../../../util/AppLink'

const serviceQuery = gql`
  query service($id: ID!) {
    service(id: $id) {
      id
      name
    }
  }
`

const useStyles = makeStyles({
  listItemText: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
})

export default function CreateAlertServiceListItem(props) {
  const { id, err } = props

  const classes = useStyles()

  const { data, loading, error: queryError } = useQuery(serviceQuery, {
    variables: {
      id,
    },
  })

  const { service } = data || {}

  if (!data && loading) return 'Loading...'
  if (queryError) return 'Error fetching data.'

  const serviceURL = '/services/' + id + '/alerts'

  return (
    <ListItem key={id} divider>
      <ListItemText disableTypography className={classes.listItemText}>
        <span>
          <Typography>
            <AppLink to={serviceURL} newTab>
              {service.name}
            </AppLink>
          </Typography>
          <Typography color='error' variant='caption'>
            {err}
          </Typography>
        </span>

        <AppLink to={serviceURL} newTab>
          <OpenInNewIcon fontSize='small' />
        </AppLink>
      </ListItemText>
    </ListItem>
  )
}

CreateAlertServiceListItem.propTypes = {
  id: p.string.isRequired,
  err: p.string,
}
