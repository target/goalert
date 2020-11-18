import React, { useState } from 'react'
import { useQuery } from '@apollo/react-hooks'
import Button from '@material-ui/core/Button'
import Divider from '@material-ui/core/Divider'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core/styles'
import gql from 'graphql-tag'
import _, { startCase, isEmpty, uniq, chain } from 'lodash'
import AdminSection from './AdminSection'
import AdminDialog from './AdminDialog'
import PageActions from '../util/PageActions'
import ExpandMoreIcon from '@material-ui/icons/ExpandMore'
import { Form } from '../forms'
import {
  InputAdornment,
  TextField,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Chip,
} from '@material-ui/core'
import CopyText from '../util/CopyText'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import { ConfigValue, ConfigHint } from '../../schema'

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
  accordionDetails: {
    padding: 0,
  },
  form: {
    width: '100%',
  },
  saveDisabled: {
    color: 'rgba(255, 255, 255, 0.5)',
  },
  heading: {
    fontSize: '1.1rem',
    flexBasis: '33.33%',
    flexShrink: 0,
  },
  secondaryHeading: {
    fontSize: theme.typography.pxToRem(15),
    color: theme.palette.text.secondary,
    flexGrow: 1,
  },
  changeChip: {
    justifyContent: 'flex-end',
  },
}))

interface ConfigValues {
  [id: string]: string
}

function formatHeading(s = ''): string {
  return startCase(s)
    .replace(/\bTwo Way\b/, 'Two-Way')
    .replace('Disable V 1 Graph QL', 'Disable V1 GraphQL')
    .replace('Git Hub', 'GitHub')
    .replace(/R Ls\b/, 'RLs') // fix usages of `URLs`
}

export default function AdminConfig(): JSX.Element {
  const classes = useStyles()
  const [confirm, setConfirm] = useState(false)
  const [values, setValues] = useState({})
  const [section, setSection] = useState(false as false | string)

  const { data, loading, error } = useQuery(query)

  if (error) {
    return <GenericError error={error.message} />
  }

  if (loading && !data) {
    return <Spinner />
  }

  const configValues: ConfigValue[] = data.config

  const updateValue = (id: string, value: null | string): void => {
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
    configValues.map((f: ConfigValue) => f.id.split('.')[0]),
  ) as string[]

  const hintGroups = chain(data.configHints)
    .groupBy((f: ConfigHint) => f.id.split('.')[0])
    .value()

  const hintName = (id: string): string => startCase(id.split('.')[1])

  const handleExpandChange = (id: string) => () =>
    setSection(id === section ? false : id)

  const hasEnable = (sectionID: string): boolean =>
    configValues.some((v) => v.id === sectionID + '.Enable')

  const isEnabled = (sectionID: string): boolean =>
    configValues.find((v) => v.id === sectionID + '.Enable')?.value === 'true'

  const changeCount = (id: string): number =>
    _.keys(values).filter((key) => key.startsWith(id + '.')).length

  return (
    <React.Fragment>
      {groups.map((groupID: string, index: number) => (
        <Accordion
          key={groupID}
          expanded={section === groupID}
          onChange={handleExpandChange(groupID)}
        >
          <AccordionSummary
            aria-expanded={section === groupID}
            aria-controls={`accordion-sect-${groupID}`}
            id={`accordion-${groupID}`}
            expandIcon={<ExpandMoreIcon />}
          >
            <Typography
              component='h2'
              variant='subtitle1'
              className={classes.heading}
            >
              {formatHeading(groupID)}
            </Typography>
            <Typography className={classes.secondaryHeading}>
              {hasEnable(groupID) &&
                (isEnabled(groupID) ? 'Enabled' : 'Disabled')}
            </Typography>
            {(changeCount(groupID) && (
              <Chip
                className={classes.changeChip}
                size='small'
                label={`${changeCount(groupID)} unsaved change${
                  changeCount(groupID) === 1 ? '' : 's'
                }`}
              />
            )) ||
              null}
          </AccordionSummary>
          <Divider />
          <AccordionDetails
            id={`accordion-sect-${groupID}`}
            aria-labelledby={`accordion-${groupID}`}
            className={classes.accordionDetails}
            role='region'
          >
            <Form className={classes.form}>
              <AdminSection
                value={values}
                onChange={(id: string, value: null | string) =>
                  updateValue(id, value)
                }
                fields={configValues
                  .filter(
                    (f: ConfigValue) => f.id.split('.')[0] === groups[index],
                  )
                  .map((f: ConfigValue) => ({
                    id: f.id,
                    label: formatHeading(_.last(f.id.split('.'))),
                    description: f.description,
                    password: f.password,
                    type: f.type,
                    value: f.value,
                  }))}
              />
              {hintGroups[groupID] &&
                hintGroups[groupID].map((h: ConfigHint) => (
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
                ))}
            </Form>
          </AccordionDetails>
        </Accordion>
      ))}
      {renderPageActions()}
      {confirm && (
        <AdminDialog
          mutation={mutation}
          values={configValues}
          fieldValues={values}
          onClose={() => setConfirm(false)}
          onComplete={() => {
            setValues({})
            setConfirm(false)
          }}
        />
      )}
    </React.Fragment>
  )
}
