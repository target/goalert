import React, { useRef, useState } from 'react'
import { PropTypes as p } from 'prop-types'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardHeader from '@mui/material/CardHeader'
import Dialog from '@mui/material/Dialog'
import Typography from '@mui/material/Typography'
import { Add } from '@mui/icons-material'
import { gql, useMutation } from '@apollo/client'
import FlatList from '../lists/FlatList'
import CreateFAB from '../lists/CreateFAB'
import PolicyStepCreateDialog from './PolicyStepCreateDialog'
import { useResetURLParams, useURLParam } from '../actions'
import DialogTitleWrapper from '../dialogs/components/DialogTitleWrapper'
import DialogContentError from '../dialogs/components/DialogContentError'
import { policyStepsQuery } from './PolicyStepsQuery'
import { useIsWidthDown } from '../util/useWidth'
import { reorderList } from '../rotations/util'
import PolicyStepEditDialog from './PolicyStepEditDialog'
import PolicyStepDeleteDialog from './PolicyStepDeleteDialog'
import OtherActions from '../util/OtherActions'
import { getStepNumber, renderChips, renderDelayMessage } from './stepUtil'

const mutation = gql`
  mutation UpdateEscalationPolicyMutation(
    $input: UpdateEscalationPolicyInput!
  ) {
    updateEscalationPolicy(input: $input)
  }
`

export default function PolicyStepsCard(props) {
  const { escalationPolicyID, repeat, steps = [] } = props

  const isMobile = useIsWidthDown('md')
  const stepNumParam = 'createStep'
  const [createStep, setCreateStep] = useURLParam(stepNumParam, false)
  const resetCreateStep = useResetURLParams(stepNumParam)

  const oldID = useRef(null)
  const oldIdx = useRef(null)
  const newIdx = useRef(null)

  const [lastSwap, setLastSwap] = useState([])

  const [error, setError] = useState(null)

  const [editStepID, setEditStepID] = useURLParam('editStep', '')
  const resetEditStep = useResetURLParams('editStep')
  const [deleteStep, setDeleteStep] = useState('')

  function arrayMove(arr) {
    const el = arr[oldIdx.current]
    arr.splice(oldIdx.current, 1)
    arr.splice(newIdx.current, 0, el)
  }

  function onStepUpdate(cache, data) {
    // mutation returns true on a success
    if (
      !data.updateEscalationPolicy ||
      oldIdx.current == null ||
      newIdx.current == null
    ) {
      return
    }

    // variables for query to read/write from the cache
    const variables = {
      id: escalationPolicyID,
    }

    // get the current state of the steps in the cache
    const { escalationPolicy } = cache.readQuery({
      query: policyStepsQuery,
      variables,
    })

    // get steps from cache
    const steps = escalationPolicy.steps.slice()

    // if optimistic cache update was successful, return out
    if (steps[newIdx.current].id === oldID.current) return

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

  const [updateEscalationPolicy] = useMutation(mutation, {
    onCompleted: () => {
      oldID.current = null
      oldIdx.current = null
      newIdx.current = null
    },
    onError: (err) => setError(err),
    update: (cache, { data }) => onStepUpdate(cache, data),
    optimisticResponse: { updateEscalationPolicy: true },
  })

  function onReorder(oldIndex, newIndex) {
    setLastSwap(lastSwap.concat({ oldIndex, newIndex }))

    const updatedStepIDs = reorderList(
      steps.map((step) => step.id),
      oldIndex,
      newIndex,
    )

    return updateEscalationPolicy({
      variables: {
        input: {
          id: escalationPolicyID,
          stepIDs: updatedStepIDs,
        },
      },
    })
  }

  function renderRepeatText() {
    if (!steps.length) {
      return null
    }

    let text = ''
    if (repeat === 0) text = 'Do not repeat'
    else if (repeat === 1) text = 'Repeat once'
    else text = `Repeat ${repeat} times`

    return (
      <Typography variant='subtitle1' component='p' sx={{ pl: 2, pb: 2 }}>
        {text}
      </Typography>
    )
  }

  const { message: errMsg } = error || {}

  return (
    <React.Fragment>
      {isMobile && (
        <CreateFAB onClick={() => setCreateStep(true)} title='Create Step' />
      )}
      {createStep && (
        <PolicyStepCreateDialog
          escalationPolicyID={escalationPolicyID}
          onClose={resetCreateStep}
        />
      )}
      <Card>
        <CardHeader
          title='Escalation Steps'
          component='h3'
          sx={{ paddingBottom: 0, margin: 0 }}
          action={
            !isMobile && (
              <Button
                variant='contained'
                onClick={() => setCreateStep(true)}
                startIcon={<Add />}
              >
                Create Step
              </Button>
            )
          }
        />
        <FlatList
          emptyMessage='No steps currently on this Escalation Policy'
          headerNote='Notify the following:'
          items={steps.map((step) => ({
            id: step.id,
            disableTypography: true,
            title: (
              <Typography component='h4' variant='subtitle1' sx={{ pb: 1 }}>
                <b>Step #{getStepNumber(step.id, steps)}:</b>
              </Typography>
            ),
            subText: (
              <React.Fragment>
                {renderChips(step)}
                {renderDelayMessage(steps, step, repeat)}
              </React.Fragment>
            ),
            secondaryAction: (
              <OtherActions
                actions={[
                  {
                    label: 'Edit',
                    onClick: () => setEditStepID(step.id),
                  },
                  {
                    label: 'Delete',
                    onClick: () => setDeleteStep(step),
                  },
                ]}
              />
            ),
          }))}
          onReorder={onReorder}
        />
        {renderRepeatText()}
      </Card>
      <Dialog open={Boolean(error)} onClose={() => setError(null)}>
        <DialogTitleWrapper
          fullScreen={useIsWidthDown('md')}
          title='An error occurred'
        />
        <DialogContentError error={errMsg} />
      </Dialog>
      {editStepID && (
        <PolicyStepEditDialog
          escalationPolicyID={escalationPolicyID}
          onClose={resetEditStep}
          step={steps.filter((step) => step.id === editStepID)[0]}
        />
      )}
      {deleteStep && (
        <PolicyStepDeleteDialog
          escalationPolicyID={escalationPolicyID}
          onClose={() => setDeleteStep(false)}
          stepID={deleteStep}
        />
      )}
    </React.Fragment>
  )
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
