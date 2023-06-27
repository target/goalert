import React, { useState, useEffect, useMemo } from 'react'
import { useQuery, useMutation, gql } from 'urql'
import { Button } from '@mui/material'
import { DEBOUNCE_DELAY } from '../../config'
import { transitionStyles } from '../../util/Transitions'

const query = gql`
  query AlertFeedbackQuery($id: Int!) {
    alert(id: $id) {
      id
      feedback {
        note
      }
    }
  }
`

const mutation = gql`
  mutation UpdateFeedbackMutation($input: UpdateAlertFeedbackInput!) {
    updateAlertFeedback(input: $input)
  }
`

interface AlertFeedbackProps {
  alertID: number
}

export default function AlertFeedback(props: AlertFeedbackProps): JSX.Element {
  const { alertID } = props
  const [cacheCount, setCacheCount] = useState(0) // reset cache on tick

  // stable query reference
  const context = useMemo(
    () => ({ additionalTypenames: ['Feedback'] }),
    [cacheCount],
  )
  const [{ data, fetching, error }] = useQuery({
    query,
    context,
    variables: {
      id: alertID,
    },
    requestPolicy: 'cache-first',
  })
  const [note, setNote] = useState(data?.alert?.feedback?.note ?? '')
  const [mutationStatus, commit] = useMutation(mutation)

  // Debounce setting note
  useEffect(() => {
    const t = setTimeout(() => {
      commit({
        input: {
          alertID,
          note,
        },
      })
    }, DEBOUNCE_DELAY)

    return () => clearTimeout(t)
  }, [note])

  useEffect(() => {
    setNote(data?.alert?.feedback?.note ?? '')
  }, [data?.alert?.feedback?.note])

  return <Button>Problem?</Button>
}
