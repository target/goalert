import React from 'react'
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
import { Mutation } from 'react-apollo'
import gql from 'graphql-tag'
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
  mutation($input: CreateServiceInput!) {
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
export default class WizardRouter extends React.PureComponent {
  state = {
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
          startDate: DateTime.local()
            .startOf('day')
            .toISO(),
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
          startDate: DateTime.local()
            .startOf('day')
            .toISO(),
          type: 'never',
          favorite: true,
        },
        followTheSunRotation: {
          enable: 'no',
          users: [],
          timeZone: null,
        },
      },
    },
  }

  /*
   * Called when submitting the entire form. The initial
   * mutation is to create a service, which will have children
   * variables that will in turn also create the proper targets.
   *
   * e.g. createService: { newEscalationPolicy: {...} }
   */
  submit = (e, isValid, commit) => {
    e.preventDefault() // prevents reloading
    if (!isValid) return

    const variables = {
      input: {
        ...getService(this.state.value),
        newEscalationPolicy: {
          ...getEscalationPolicy(this.state.value),
          steps: this.getSteps(),
        },
      },
    }

    commit({ variables })
      .then(() => {
        this.setState({ complete: true })
      })
      .catch(err => {
        const generalErrors = nonFieldErrors(err)
        const graphqlErrors = fieldErrors(err).map(error => {
          const name = error.field
            .split('.')
            .pop() // get last occurrence
            .replace(/([A-Z])/g, ' $1') // insert a space before all caps
            .replace(/^./, str => str.toUpperCase()) // uppercase the first character

          return `${name}: ${error.message}`
        })

        const errors = generalErrors.concat(graphqlErrors)

        if (errors.length) {
          this.setState({
            errorMessage: errors.map(e => e.message || e).join('\n'),
          })
        }
      })
  }

  /*
   * Get steps for the EP
   *
   * Handles not returning a second step if the secondary
   * schedule is not enabled in the form.
   */
  getSteps = () => {
    const value = this.state.value
    const secondary = value.secondarySchedule.enable === 'yes'
    const steps = []

    const step = key => ({
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

  onDialogClose = data => {
    if (data && data.createService) {
      return this.setState({ redirect: true })
    }

    this.setState({ complete: false, errorMessage: null }, () => {
      window.scrollTo(0, 0)
    })
  }

  render() {
    const { complete, errorMessage, redirect } = this.state

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
                  onSubmit={(e, isValid) => this.submit(e, isValid, commit)}
                  disabled={loading}
                >
                  <CardContent>
                    <WizardForm
                      disabled={status.loading}
                      errors={fieldErrors(error)}
                      value={this.state.value}
                      onChange={value => this.setState({ value })}
                    />
                  </CardContent>
                  <CardActions className={this.props.classes.cardActions}>
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
                  isWidthDown('md', this.props.width) && !errorMessage
                }
                open={complete || Boolean(errorMessage)}
                onClose={() => this.onDialogClose(data)}
              >
                <DialogTitleWrapper
                  fullScreen={isWidthDown('md', this.props.width)}
                  title={complete ? 'Success!' : 'An error occurred'}
                />
                {this.renderSubmittedContent()}
                <DialogActions>
                  <Button
                    onClick={() => this.onDialogClose(data)}
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

  renderSubmittedContent() {
    const { complete, errorMessage, value } = this.state

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
}
