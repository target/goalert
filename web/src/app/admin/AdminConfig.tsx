import React, { useState } from 'react'
import { useQuery, gql } from 'urql'
import Button from '@mui/material/Button'
import ButtonGroup from '@mui/material/ButtonGroup'
import Divider from '@mui/material/Divider'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'
import _, { startCase, isEmpty, uniq, chain } from 'lodash'
import AdminSection from './AdminSection'
import AdminDialog from './AdminDialog'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import { Form } from '../forms'
import {
  InputAdornment,
  TextField,
  Accordion,
  AccordionSummary,
  AccordionDetails,
  Chip,
} from '@mui/material'
import CopyText from '../util/CopyText'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import { ConfigValue, ConfigHint } from '../../schema'
import SlackActions from './SlackActions'

const query = gql`
  query getConfig {
    config(all: true) {
      id
      description
      password
      type
      value
      deprecated
    }
    configHints {
      id
      value
    }
  }
`

const useStyles = makeStyles((theme: Theme) => ({
  accordionDetails: {
    padding: 0,
    display: 'block',
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
    .replace('Enable V 1 Graph QL', 'Enable V1 GraphQL')
    .replace('Git Hub', 'GitHub')
    .replace(/R Ls\b/, 'RLs') // fix usages of `URLs`
}

export default function AdminConfig(): React.ReactNode {
  const classes = useStyles()
  const [confirm, setConfirm] = useState(false)
  const [values, setValues] = useState({})
  const [section, setSection] = useState(false as false | string)

  const [{ data, fetching, error }] = useQuery({ query })

  if (error) {
    return <GenericError error={error.message} />
  }

  if (fetching && !data) {
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
    <Grid container spacing={2}>
      <Grid item xs={12} container justifyContent='flex-end'>
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
          value={values}
          onClose={() => setConfirm(false)}
          onComplete={() => {
            setValues({})
            setConfirm(false)
          }}
        />
      )}

      <Grid item xs={12}>
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
                      deprecated: f.deprecated,
                    }))}
                />
              </Form>
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
                          <CopyText value={h.value} placement='left' asURL />
                        </InputAdornment>
                      ),
                    }}
                    fullWidth
                  />
                ))}
              {groupID === 'Slack' && <SlackActions />}
            </AccordionDetails>
          </Accordion>
        ))}
      </Grid>
    </Grid>
  )
}
