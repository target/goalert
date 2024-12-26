import React, { useState } from 'react'
import { gql, useQuery, useMutation } from 'urql'

import { fieldErrors, nonFieldErrors } from '../util/errutil'
import Checkbox from '@mui/material/Checkbox'
import FormControlLabel from '@mui/material/FormControlLabel'
import FormControl from '@mui/material/FormControl'
import FormHelperText from '@mui/material/FormHelperText'
import Typography from '@mui/material/Typography'
import FormDialog from '../dialogs/FormDialog'
import Spinner from '../loading/components/Spinner'
import _ from 'lodash'
import { useLocation } from 'wouter'

function DeleteForm(props: {
  epName: string
  error: string | undefined
  value: boolean
  onChange: (deleteEP: boolean) => void
}): React.JSX.Element {
  return (
    <FormControl error={Boolean(props.error)} style={{ width: '100%' }}>
      <FormControlLabel
        control={
          <Checkbox
            checked={props.value}
            onChange={(e) => props.onChange(e.target.checked)}
            value='delete-escalation-policy'
          />
        }
        label={
          <React.Fragment>
            Also delete escalation policy: {props.epName}
          </React.Fragment>
        }
      />
      <FormHelperText>{props.error}</FormHelperText>
    </FormControl>
  )
}

const query = gql`
  query service($id: ID!) {
    service(id: $id) {
      id
      name
      ep: escalationPolicy {
        id
        name
      }
    }
  }
`
const mutation = gql`
  mutation delete($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default function ServiceDeleteDialog(props: {
  serviceID: string
  onClose: () => void
}): React.JSX.Element {
  const [, navigate] = useLocation()
  const [deleteEP, setDeleteEP] = useState<boolean>(true)
  const [{ data, ...dataStatus }] = useQuery({
    query,
    variables: { id: props.serviceID },
  })
  const input = [{ type: 'service', id: props.serviceID }]
  const [deleteServiceStatus, deleteService] = useMutation(mutation)

  const epID = _.get(data, 'service.ep.id')
  const epName = _.get(
    data,
    'service.ep.name',
    <Spinner text='fetching policy...' />,
  )

  if (deleteEP) {
    input.push({
      type: 'escalationPolicy',
      id: epID,
    })
  }

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={
        <Typography component='h2' variant='subtitle1'>
          This will delete the service:{' '}
          {_.get(data, 'service.name', <Spinner text='loading...' />)}
        </Typography>
      }
      caption='Deleting a service will also delete all associated integration keys and alerts.'
      loading={deleteServiceStatus.fetching || (!data && dataStatus.fetching)}
      errors={nonFieldErrors(deleteServiceStatus.error)}
      onClose={props.onClose}
      onSubmit={() => {
        deleteService(
          { input },
          {
            additionalTypenames: ['Service'],
          },
        ).then((res) => {
          if (res.error) return
          navigate('/services')
        })
      }}
      form={
        <DeleteForm
          epName={epName}
          error={
            fieldErrors(deleteServiceStatus.error).find(
              (f) => f.field === 'escalationPolicyID',
            ) && 'Escalation policy is currently in use.'
          }
          onChange={(deleteEP) => setDeleteEP(deleteEP)}
          value={deleteEP}
        />
      }
    />
  )
}
