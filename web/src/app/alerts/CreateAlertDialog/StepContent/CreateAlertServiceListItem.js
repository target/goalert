import React from 'react'
import p from 'prop-types'
import { useQuery } from 'react-apollo'
import {
  ListItem,
  ListItemText,
  Typography,
  Link,
  makeStyles,
} from '@material-ui/core'
import OpenInNewIcon from '@material-ui/icons/OpenInNew'
import { useSelector } from 'react-redux'

import gql from 'graphql-tag'
import { absURLSelector } from '../../../selectors'

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
  const absURL = useSelector(absURLSelector)

  const { data, loading, error: queryError } = useQuery(serviceQuery, {
    variables: {
      id,
    },
  })

  const { service } = data || {}

  if (loading) return 'Loading...'
  if (queryError) return 'Error fetching data.'

  const serviceURL = absURL('/services/' + id + '/alerts')

  return (
    <ListItem key={id} divider>
      <ListItemText disableTypography className={classes.listItemText}>
        <span>
          <Typography>
            <Link href={serviceURL} target='_blank' rel='noopener noreferrer'>
              {service.name}
            </Link>
          </Typography>
          <Typography color='error' variant='caption'>
            {err}
          </Typography>
        </span>

        <a href={serviceURL} target='_blank' rel='noopener noreferrer'>
          <OpenInNewIcon fontSize='small' />
        </a>
      </ListItemText>
    </ListItem>
  )
}

CreateAlertServiceListItem.propTypes = {
  id: p.string.isRequired,
  err: p.string,
}
