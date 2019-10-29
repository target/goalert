import React from 'react'
import { DialogActions, Button } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import gql from 'graphql-tag'
import { useMutation } from 'react-apollo'
import { fieldAlias, mergeFields, mapInputVars } from '../../util/graphql'

const baseMutation = gql`
  mutation CreateAlertMutation($input: CreateAlertInput!) {
    createAlert(input: $input) {
      id
    }
  }
`
/*
TRANSFORM TO
mutation {
  alias0: createAlert(
    input: {
      summary: "mysummary"
      details: "mydetails"
      serviceID: "292a2d74-213c-440d-b28b-7606b0dd02f1"
    }
  ) {
    id
  },
  alias1: createAlert(
    input: {
      summary: "mysummary"
      details: "mydetails"
      serviceID: "djf82d74-213c-we0d-b28b-fefv0dd02f1"
    }
  ) {
    id
  },

  ...

}
*/

const useStyles = makeStyles(theme => ({
  button: {
    marginRight: theme.spacing(1),
  },
}))

const nextIsDisabled = (activeStep, formFields) => {
  switch (activeStep) {
    case 0:
      return !(formFields.summary && formFields.details)
    case 1:
      return formFields.selectedServices.length === 0
    default:
      return false
  }
}

export default props => {
  const { activeStep, formFields, setActiveStep, steps } = props
  const classes = useStyles()

  const handleNext = () => {
    setActiveStep(prevActiveStep => prevActiveStep + 1)
  }

  const handleBack = () => {
    setActiveStep(prevActiveStep => prevActiveStep - 1)
  }

  const onLastStep = () => activeStep === steps.length - 1

  const getAliasedMutation = (mutation, index) =>
    mapInputVars(fieldAlias(mutation, 'data' + index), {
      input: 'input' + index,
    })

  const makeCreateAlerts = () => {
    // 1. build mutation
    let m = getAliasedMutation(baseMutation, 0)

    for (let i = 1; i < formFields.selectedServices.length; i++) {
      m = mergeFields(m, getAliasedMutation(baseMutation, i))
    }

    console.log(m)

    // 2. build variables
    /*
    {
      input1: {
        summary: 'my summary',
        details: 'my details',
        serviceID: 'wefwe-ewf-wef-wef'
      },
      input2: {
        summary: 'my summary',
        details: 'my details',
        serviceID: 'rtyt-rty-rty-rty'
      },
      ...
    }
    */
    let variables = {}
    formFields.selectedServices.forEach((ss, i) => {
      variables[`input${i}`] = {
        summary: formFields.summary,
        details: formFields.details,
        serviceID: ss,
      }
    })

    // 3. execute mutation with variables
    return useMutation(m, {
      variables,
    })
  }

  const [createAlerts] = makeCreateAlerts()

  return (
    <DialogActions>
      <Button
        disabled={activeStep === 0}
        onClick={handleBack}
        className={classes.button}
      >
        Back
      </Button>

      <Button
        variant='contained'
        color='primary'
        onClick={onLastStep() ? () => createAlerts() : handleNext}
        className={classes.button}
        disabled={nextIsDisabled(activeStep, formFields)}
      >
        {onLastStep() ? 'Submit' : 'Next'}
      </Button>
    </DialogActions>
  )
}
