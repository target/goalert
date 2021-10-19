import React, { useState } from 'react'
import Button from '@material-ui/core/Button'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import CardActions from '@material-ui/core/CardActions'
import Dialog from '@material-ui/core/Dialog'
import DialogContent from '@material-ui/core/DialogContent'
import DialogContentText from '@material-ui/core/DialogContentText'
import DialogActions from '@material-ui/core/DialogActions'
import { Redirect } from 'react-router-dom'
import { makeStyles } from '@material-ui/core'
import { DateTime } from 'luxon'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import WizardForm from './WizardForm'
import LoadingButton from '../loading/components/LoadingButton'
import { Mutation } from '@apollo/client/react/components'
import { gql } from '@apollo/client'
import { Form } from '../forms'
import {
  getService,
  getEscalationPolicy,
  getSchedule,
  getScheduleTargets,
} from './util'
import DialogTitleWrapper from '../dialogs/components/DialogTitleWrapper'
import DialogContentError from '../dialogs/components/DialogContentError'
import { useIsWidthDown } from '../util/useWidth'

const mutation = gql`
  mutation ($input: CreateServiceInput!) {
    createService(input: $input) {
      id
      name
      description
      escalationPolicyID
    }
  }
`

const useStyles = makeStyles(() => ({
  cardActions: {
    justifyContent: 'flex-end',
  },
}))

export default function WizardRouter() {
  const classes = useStyles()
  const fullScreen = useIsWidthDown('md')
  const [errorMessage, setErrorMessage] = useState(null)
  const [complete, setComplete] = useState(false)
  const [redirect, setRedirect] = useState(false)
  const [value, setValue] = useState({
    teamName: '',
    delayMinutes: '',
    repeat: '',
    key: null,
    primarySchedule: {
      timeZone: null,
      users: [],
      rotation: {
        startDate: DateTime.local().startOf('day').toISO(),
        type: 'never',
        favorite: true,
      },
      followTheSunRotation: {
        enable: 'no',
        users: [],
        timeZone: null,
      },
    },
    secondarySchedule: {
      enable: 'no',
      timeZone: null,
      users: [],
      rotation: {
        startDate: DateTime.local().startOf('day').toISO(),
        type: 'never',
        favorite: true,
      },
      followTheSunRotation: {
        enable: 'no',
        users: [],
        timeZone: null,
      },
    },
  })

  /*
   * Get steps for the EP
   *
   * Handles not returning a second step if the secondary
   * schedule is not enabled in the form.
   */
  const getSteps = () => {
    const secondary = value.secondarySchedule.enable === 'yes'
    const steps = []

    const step = (key) => ({
      delayMinutes: value.delayMinutes,
      newSchedule: {
        ...getSchedule(key, value),
        targets: getScheduleTargets(key, value),
      },
    })

    // push first step
    steps.push(step('primarySchedule'))

    // push second step
    if (secondary) steps.push(step('secondarySchedule'))

    return steps
  }

  /*
   * Called when submitting the entire form. The initial
   * mutation is to create a service, which will have children
   * variables that will in turn also create the proper targets.
   *
   * e.g. createService: { newEscalationPolicy: {...} }
   */
  const submit = (e, isValid, commit) => {
    e.preventDefault() // prevents reloading
    if (!isValid) return

    const variables = {
      input: {
        ...getService(value),
        newEscalationPolicy: {
          ...getEscalationPolicy(value),
          steps: getSteps(),
        },
      },
    }

    commit({ variables })
      .then(() => {
        setComplete(true)
      })
      .catch((err) => {
        const generalErrors = nonFieldErrors(err)
        const graphqlErrors = fieldErrors(err).map((error) => {
          const name = error.field
            .split('.')
            .pop() // get last occurrence
            .replace(/([A-Z])/g, ' $1') // insert a space before all caps
            .replace(/^./, (str) => str.toUpperCase()) // uppercase the first character

          return `${name}: ${error.message}`
        })

        const errors = generalErrors.concat(graphqlErrors)

        if (errors.length) {
          setErrorMessage(errors.map((e) => e.message || e).join('\n'))
        }
      })
  }

  const onDialogClose = (data) => {
    if (data && data.createService) {
      return setRedirect(true)
    }

    setComplete(false)
    setErrorMessage(null)
    window.scrollTo(0, 0)
  }

  function renderSubmittedContent() {
    if (complete) {
      return (
        <DialogContent>
          <DialogContentText>
            You can search for each of the targets you have created by the name
            "{value.teamName}" within GoAlert. Upon closing this dialog, you
            will be routed to your newly created service.
          </DialogContentText>
        </DialogContent>
      )
    }
    if (errorMessage) {
      return <DialogContentError error={errorMessage} />
    }
  }

  return (
    <Mutation mutation={mutation}>
      {(commit, { data, error, loading }) => {
        if (redirect && data && data.createService) {
          return <Redirect push to={`/services/${data.createService.id}`} />
        }

        return (
          <React.Fragment>
            <Card>
              <Form
                onSubmit={(e, isValid) => submit(e, isValid, commit)}
                disabled={loading}
              >
                <CardContent>
                  <WizardForm
                    disabled={status.loading}
                    errors={fieldErrors(error)}
                    value={value}
                    onChange={(value) => setValue(value)}
                  />
                </CardContent>
                <CardActions className={classes.cardActions}>
                  <LoadingButton
                    attemptCount={fieldErrors(error).length ? 1 : 0}
                    buttonText='Submit'
                    color='primary'
                    loading={loading}
                    type='submit'
                  />
                </CardActions>
              </Form>
            </Card>
            <Dialog
              fullScreen={fullScreen && !errorMessage}
              open={complete || Boolean(errorMessage)}
              onClose={() => onDialogClose(data)}
            >
              <DialogTitleWrapper
                fullScreen={fullScreen}
                title={complete ? 'Success!' : 'An error occurred'}
              />
              {renderSubmittedContent()}
              <DialogActions>
                <Button onClick={() => onDialogClose(data)} color='primary'>
                  Close
                </Button>
              </DialogActions>
            </Dialog>
          </React.Fragment>
        )
      }}
    </Mutation>
  )
}
