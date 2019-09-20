import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { useMutation } from '@apollo/react-hooks'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import Dialog from '@material-ui/core/Dialog'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core/styles/index'
import gql from 'graphql-tag'
import PolicyStep from './PolicyStep'
import DialogTitleWrapper from '../dialogs/components/DialogTitleWrapper'
import DialogContentError from '../dialogs/components/DialogContentError'
import { policyStepsQuery } from './PolicyStepsQuery'
import FlatList from '../lists/FlatList'

const useStyles = makeStyles({
  headerEl: {
    color: 'black',
  },
})

const updateEPMutation = gql`
  mutation UpdateEscalationPolicyMutation(
    $input: UpdateEscalationPolicyInput!
  ) {
    updateEscalationPolicy(input: $input)
  }
`

export default function PolicyStepsCard(props) {
  let oldID = null
  let oldIdx = null
  let newIdx = null
  let stepIDs = props.steps.map(step => step.id)

  const classes = useStyles()
  const [showErrorDialog, setShowErrorDialog] = useState(false)
  const [updateEP, updateEPStatus] = useMutation(updateEPMutation, {
    onCompleted: () => {
      oldID = null
      oldIdx = null
      newIdx = null
    },
    onError: () => setShowErrorDialog(true),
    optimisticResponse: {
      updateEscalationPolicy: true,
    },
    update: (cache, { data }) => updateCache(cache, data),
    variables: {
      input: {
        id: props.escalationPolicyID,
        stepIDs,
      },
    },
  })

  function arrayMove(arr) {
    const el = arr[oldIdx]
    arr.splice(oldIdx, 1)
    arr.splice(newIdx, 0, el)
  }

  /*
   * Executes on drag end. Once the mutation completes
   * successfully, updateCache will be called to update
   * the UI with the correct data.
   */
  function onReorder(result) {
    // dropped outside the list
    if (!result.destination) {
      return
    }

    oldID = result.draggableId
    oldIdx = stepIDs.indexOf(oldID)
    newIdx = result.destination.index

    // re-order sids array
    arrayMove(stepIDs)

    // call mutation
    return updateEP()
  }

  function updateCache(cache, data) {
    // mutation returns true on a success
    if (!data.updateEscalationPolicy || oldIdx == null || newIdx == null) {
      return
    }

    // variables for query to read/write from the cache
    const variables = {
      id: props.escalationPolicyID,
    }

    // get the current state of the steps in the cache
    const { escalationPolicy } = cache.readQuery({
      query: policyStepsQuery,
      variables,
    })

    // get steps from cache
    const steps = escalationPolicy.steps.slice()

    // if optimistic cache update was successful, return out
    if (steps[newIdx].id === oldID) return

    // re-order escalationPolicy.steps array
    arrayMove(steps)

    // write new steps order to cache
    cache.writeQuery({
      query: policyStepsQuery,
      variables,
      data: {
        escalationPolicy: {
          ...escalationPolicy,
          steps,
        },
      },
    })
  }

  // main render return
  return (
    <React.Fragment>
      <Card>
        <CardContent>
          <Typography variant='h5' component='h3'>
            Escalation Steps
          </Typography>
          {renderStepsList()}
          {renderRepeatText()}
        </CardContent>
      </Card>
      <Dialog open={showErrorDialog} onClose={() => setShowErrorDialog(false)}>
        <DialogTitleWrapper title='An error occurred' />
        <DialogContentError
          error={updateEPStatus.error && updateEPStatus.error.message}
        />
      </Dialog>
    </React.Fragment>
  )

  function renderStepsList() {
    const { escalationPolicyID, repeat, steps } = props

    const headerEl = (
      <Typography
        component='p'
        variant='subtitle1'
        className={classes.headerEl}
      >
        Notify the following:
      </Typography>
    )

    return (
      <React.Fragment>
        <FlatList
          data-cy='steps-list'
          emptyMessage='No steps currently on this Escalation Policy'
          headerNote={steps.length ? headerEl : null}
          onReorder={onReorder}
          items={steps.map((step, idx) => ({
            id: step.id,
            el: (
              <PolicyStep
                key={idx}
                escalationPolicyID={escalationPolicyID}
                index={idx}
                repeat={repeat}
                step={step}
                steps={steps}
              />
            ),
          }))}
        />
      </React.Fragment>
    )
  }

  function renderRepeatText() {
    const { repeat, steps } = props

    if (!steps.length) {
      return null
    }

    let text = ''
    if (repeat === 0) text = 'Do not repeat'
    else if (repeat === 1) text = 'Repeat once'
    else text = `Repeat ${repeat} times`

    return (
      <Typography variant='subtitle1' component='p'>
        {text}
      </Typography>
    )
  }
}

PolicyStepsCard.propTypes = {
  escalationPolicyID: p.string.isRequired,
  repeat: p.number.isRequired, // # of times EP repeats escalation process
  steps: p.arrayOf(
    p.shape({
      id: p.string.isRequired,
      delayMinutes: p.number.isRequired,
      targets: p.arrayOf(
        p.shape({
          id: p.string.isRequired,
          name: p.string.isRequired,
          type: p.string.isRequired,
        }),
      ).isRequired,
    }),
  ).isRequired,
}
