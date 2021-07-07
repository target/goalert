import React, {useState} from 'react'
import Button from '@material-ui/core/Button'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import CardActions from '@material-ui/core/CardActions'
import Dialog from '@material-ui/core/Dialog'
import DialogContent from '@material-ui/core/DialogContent'
import DialogContentText from '@material-ui/core/DialogContentText'
import DialogActions from '@material-ui/core/DialogActions'
import withStyles from '@material-ui/core/styles/withStyles'
import { Redirect } from 'react-router-dom'
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
import withWidth, { isWidthDown } from '@material-ui/core/withWidth'
import DialogTitleWrapper from '../dialogs/components/DialogTitleWrapper'
import DialogContentError from '../dialogs/components/DialogContentError'

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

const styles = {
  cardActions: {
    justifyContent: 'flex-end',
  },
}

@withWidth()
@withStyles(styles)

export default function WizardRouter ({classes, width}) {
 const [state, setState] = useState({
    errorMessage: null,
    complete: false,
    redirect: false,
    // all of the potential variables to be used in the final mutation
    // objected prefaced with "enable" are optional, and won't committed
    // unless set to 'yes'
    value: {
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
    }})


    /*
   * Get steps for the EP
   *
   * Handles not returning a second step if the secondary
   * schedule is not enabled in the form.
   */
  const getSteps = () => {
    const value = state.value
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
        ...getService(state.value),
        newEscalationPolicy: {
          ...getEscalationPolicy(state.value),
          steps: getSteps(),
        },
      },
    }

    commit({ variables })
      .then(() => {
        setState({ complete: true })
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
          setState({
            errorMessage: errors.map((e) => e.message || e).join('\n'),
          })
        }
      })
  }


  const onDialogClose = (data) => {
    if (data && data.createService) {
      return setState({ redirect: true })
    }

    setState({ complete: false, errorMessage: null }, () => {
      window.scrollTo(0, 0)
    })
  }
  
    const { complete, errorMessage, redirect } = state

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
                      value={state.value}
                      onChange={(value) => setState({ value })}
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
                fullScreen={
                  isWidthDown('md', width) && !errorMessage
                }
                open={complete || Boolean(errorMessage)}
                onClose={() => onDialogClose(data)}
              >
                <DialogTitleWrapper
                  fullScreen={isWidthDown('md', width)}
                  title={complete ? 'Success!' : 'An error occurred'}
                />
                {renderSubmittedContent(state)}
                <DialogActions>
                  <Button
                    onClick={() => onDialogClose(data)}
                    color='primary'
                  >
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

  function renderSubmittedContent(state) {
    const { complete, errorMessage, value } = state

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
