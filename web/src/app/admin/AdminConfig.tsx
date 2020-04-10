import React, { useState } from 'react'
import { useQuery } from '@apollo/react-hooks'
import Button from '@material-ui/core/Button'
import Card from '@material-ui/core/Card'
import Grid from '@material-ui/core/Grid'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core/styles'
import gql from 'graphql-tag'
import { startCase, isEmpty, uniq } from 'lodash-es'
import chain from 'lodash'
import AdminSection from './AdminSection'
import AdminDialog from './AdminDialog'
import PageActions from '../util/PageActions'
import { Form } from '../forms'
import { InputAdornment, TextField } from '@material-ui/core'
import CopyText from '../util/CopyText'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'

const query = gql`
  query getConfig {
    config(all: true) {
      id
      description
      password
      type
      value
    }
    configHints {
      id
      value
    }
  }
`
const mutation = gql`
  mutation($input: [ConfigValueInput!]) {
    setConfig(input: $input)
  }
`

const useStyles = makeStyles((theme) => ({
  gridContainer: {
    [theme.breakpoints.up('md')]: {
      justifyContent: 'center',
    },
  },
  gridItem: {
    [theme.breakpoints.up('md')]: {
      maxWidth: '65%',
    },
  },
  groupTitle: {
    fontSize: '1.1rem',
  },
  saveDisabled: {
    color: 'rgba(255, 255, 255, 0.5)',
  },
}))

interface ConfigValues {
  [id: string]: string
}

export default function AdminConfig(): JSX.Element {
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

  const updateValue = (id: string, value: string): void => {
    const newVal: ConfigValues = { ...values }

    if (value === null) {
      delete newVal[id]
    } else {
      newVal[id] = value
    }

    setValues(newVal)
  }

  const renderPageActions = (): JSX.Element => {
    return (
      <PageActions>
        <Button
          color='inherit'
          data-cy='reset'
          disabled={isEmpty(values)}
          onClick={() => setValues({})}
          classes={{
            label: isEmpty(values) ? classes.saveDisabled : undefined,
          }}
        >
          Reset
        </Button>
        <Button
          color='inherit'
          data-cy='save'
          disabled={isEmpty(values)}
          onClick={() => setConfirm(true)}
          classes={{
            label: isEmpty(values) ? classes.saveDisabled : undefined,
          }}
        >
          Save
        </Button>
      </PageActions>
    )
  }

  const groups = uniq(
    data.config.map((f: { id: string }) => f.id.split('.')[0]),
  ) as string[]
  const hintGroups = chain(data.configHints)
    .groupBy((f: { id: string }) => f.id.split('.')[0])
    .value()
  const hintName = (id: string): string => startCase(id.split('.')[1])

  return (
    <div>
      <Grid container spacing={2} className={classes.gridContainer}>
        {groups.map((groupID: string, index: number) => (
          <Grid
            key={index}
            container // contains title above card/card itself
            item // for each admin config section
            xs={12}
            className={classes.gridItem}
          >
            <Grid item xs={12}>
              <Typography
                component='h2'
                variant='subtitle1'
                color='textSecondary'
                classes={{
                  subtitle1: classes.groupTitle,
                }}
              >
                {startCase(groupID).replace('Git Hub', 'GitHub')}
              </Typography>
            </Grid>
            <Grid item xs={12}>
              <Form>
                <Card>
                  <AdminSection
                    value={values}
                    onChange={(id: string, value: string) =>
                      updateValue(id, value)
                    }
                    fields={data.config
                      .filter(
                        (f: { id: string }) =>
                          f.id.split('.')[0] === groups[index],
                      )
                      .map(
                        (f: {
                          id: string
                          description: string
                          password: boolean
                          type: string
                          value: string
                        }) => ({
                          id: f.id,
                          label: startCase(
                            chain(f.id.split('.')).last(),
                          ).replace(/R Ls\b/, 'RLs'), // fix usages of `URLs`
                          description: f.description,
                          password: f.password,
                          type: f.type,
                          value: f.value,
                        }),
                      )}
                  />
                  {hintGroups[groupID] &&
                    hintGroups[groupID].map(
                      (h: { id: string; value: string }) => (
                        <TextField
                          key={h.id}
                          label={hintName(h.id)}
                          value={h.value}
                          variant='filled'
                          margin='none'
                          InputProps={{
                            endAdornment: (
                              <InputAdornment position='end'>
                                <CopyText value={h.value} placement='left' />
                              </InputAdornment>
                            ),
                          }}
                          fullWidth
                        />
                      ),
                    )}
                </Card>
              </Form>
            </Grid>
          </Grid>
        ))}
      </Grid>
      {renderPageActions()}
      {confirm && (
        <AdminDialog
          mutation={mutation}
          values={data.config}
          fieldValues={values}
          onClose={() => setConfirm(false)}
          onComplete={() => {
            setValues({})
            setConfirm(false)
          }}
        />
      )}
    </div>
  )
}
