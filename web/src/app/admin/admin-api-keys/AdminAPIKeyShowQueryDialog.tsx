import React from 'react'
import FormDialog from '../../dialogs/FormDialog'
import { gql, useQuery } from 'urql'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import { GQLAPIKey } from '../../../schema'
import { Grid, useTheme } from '@mui/material'
import CopyText from '../../util/CopyText'

// query for getting existing API Keys
const query = gql`
  query gqlAPIKeysQuery {
    gqlAPIKeys {
      id
      name
      query
    }
  }
`

export default function AdminAPIKeyShowQueryDialog(props: {
  apiKeyID: string
  onClose: (yes: boolean) => void
}): JSX.Element {
  const theme = useTheme()
  const [{ fetching, data, error }] = useQuery({
    query,
  })
  const { apiKeyID, onClose } = props

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  const key = data?.gqlAPIKeys?.find((d: GQLAPIKey) => {
    return d.id === apiKeyID
  })
  if (!key) throw new Error('API Key not found')

  return (
    <FormDialog
      title='API Key Query'
      alert
      subTitle={
        'This is the fixed query allowed for the API key: ' + key.name + '.'
      }
      loading={fetching}
      form={
        <Grid container spacing={2}>
          <Grid item xs={12}>
            <code
              style={{
                fontSize: 'large',
                color: theme.palette.getContrastText(
                  theme.palette.secondary.main,
                ),
              }}
            >
              <pre
                style={{
                  backgroundColor: theme.palette.secondary.main,
                  padding: 12,
                  borderRadius: 6,
                }}
              >
                {key.query}
              </pre>
            </code>
          </Grid>
          <Grid item xs={12}>
            <CopyText title='Copy Query' value={key.query} placement='bottom' />
          </Grid>
          <Grid item xs={12}>
            <CopyText
              title='Copy Query (as JSON)'
              value={JSON.stringify(key.query)}
              placement='bottom'
            />
          </Grid>
        </Grid>
      }
      onClose={onClose}
    />
  )
}
