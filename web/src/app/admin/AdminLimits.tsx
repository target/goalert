import React, { useState } from 'react'
import { useQuery, gql } from 'urql'
import Button from '@mui/material/Button'
import ButtonGroup from '@mui/material/ButtonGroup'
import Card from '@mui/material/Card'
import Grid from '@mui/material/Grid'
import Box from '@mui/material/Box'
import { startCase, isEmpty } from 'lodash'
import AdminDialog from './AdminDialog'
import { Form } from '../forms'
import AdminSection from './AdminSection'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'

const query = gql`
  query getLimits {
    values: systemLimits {
      id
      description
      value
    }
  }
`

const mutation = gql`
  mutation ($input: [SystemLimitInput!]!) {
    setSystemLimits(input: $input)
  }
`

interface LimitsValues {
  [id: string]: string
}

export default function AdminLimits(): React.JSX.Element {
  const [confirm, setConfirm] = useState(false)
  const [values, setValues] = useState({})

  const [{ data, fetching, error }] = useQuery({ query })

  if (error) {
    return <GenericError error={error.message} />
  }

  if (fetching && !data) {
    return <Spinner />
  }

  const updateValue = (id: string, value: null | string): void => {
    const newVal: LimitsValues = { ...values }

    if (value === null) {
      delete newVal[id]
    } else {
      newVal[id] = value
    }

    setValues(newVal)
  }

  return (
    <Box sx={{ display: 'flex', justifyContent: 'center' }}>
      <Grid
        container
        spacing={2}
        sx={(theme) => ({
          [theme.breakpoints.up('md')]: { maxWidth: '65%' },
        })}
      >
        <Grid
          item
          xs={12}
          sx={{ display: 'flex', justifyContent: 'flex-end' }}
        >
          <ButtonGroup variant='outlined'>
            <Button
              data-cy='reset'
              disabled={isEmpty(values)}
              onClick={() => setValues({})}
            >
              Reset
            </Button>
            <Button
              data-cy='save'
              disabled={isEmpty(values)}
              onClick={() => setConfirm(true)}
            >
              Save
            </Button>
          </ButtonGroup>
        </Grid>

        {confirm && (
          <AdminDialog
            query={query}
            mutation={mutation}
            value={values}
            onClose={() => setConfirm(false)}
            onComplete={() => {
              setValues({})
              setConfirm(false)
            }}
          />
        )}

        <Grid item xs={12}>
          <Form>
            <Card>
              <AdminSection
                value={values}
                onChange={(id: string, value: null | string) =>
                  updateValue(id, value)
                }
                headerNote='Set limits to -1 to disable.'
                fields={data.values.map(
                  (f: { id: string; description: string; value: number }) => ({
                    id: f.id,
                    type: 'integer',
                    description: f.description,
                    value: f.value.toString(),
                    label: startCase(f.id.replace(/([a-z])([A-Z])/g, '$1 $2')),
                    password: false,
                  }),
                )}
              />
            </Card>
          </Form>
        </Grid>
      </Grid>
    </Box>
  )
}
