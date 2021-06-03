import React, { useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import Button from '@material-ui/core/Button'
import ButtonGroup from '@material-ui/core/ButtonGroup'
import Card from '@material-ui/core/Card'
import Grid from '@material-ui/core/Grid'
import { makeStyles } from '@material-ui/core/styles'
import { startCase, isEmpty } from 'lodash'
import AdminDialog from './AdminDialog'
import { Form } from '../forms'
import AdminSection from './AdminSection'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'

const query = gql`
  query getLimits {
    systemLimits {
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

const useStyles = makeStyles((theme) => ({
  actionsContainer: {
    display: 'flex',
    justifyContent: 'flex-end',
  },
  gridContainer: {
    [theme.breakpoints.up('md')]: {
      maxWidth: '65%',
    },
  },
  groupTitle: {
    fontSize: '1.1rem',
  },
  pageContainer: {
    display: 'flex',
    justifyContent: 'center',
  },
}))

interface LimitsValues {
  [id: string]: string
}

export default function AdminLimits(): JSX.Element {
  const classes = useStyles()
  const [confirm, setConfirm] = useState(false)
  const [values, setValues] = useState({})

  const { data, loading, error } = useQuery(query)

  if (error) {
    return <GenericError error={error.message} />
  }

  if (loading && !data) {
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
    <div className={classes.pageContainer}>
      <Grid container spacing={2} className={classes.gridContainer}>
        <Grid item xs={12} className={classes.actionsContainer}>
          <ButtonGroup color='primary' variant='outlined'>
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
            mutation={mutation}
            values={data.systemLimits}
            fieldValues={values}
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
                fields={data.systemLimits.map(
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
    </div>
  )
}
