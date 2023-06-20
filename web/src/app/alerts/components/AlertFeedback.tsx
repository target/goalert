import React, { useState, useEffect, useMemo } from 'react'
import { useQuery, useMutation, gql } from 'urql'
import { Grid, Grow, IconButton, Tooltip, TextField } from '@mui/material'
import {
  ThumbUp,
  ThumbDown,
  ThumbUpOutlined,
  ThumbDownOutlined,
  Info,
} from '@mui/icons-material'
import { DEBOUNCE_DELAY } from '../../config'
import { transitionStyles } from '../../util/Transitions'

const query = gql`
  query AlertFeedbackQuery($id: Int!) {
    alert(id: $id) {
      id
      feedback {
        sentiment
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
  const classes = transitionStyles()
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

  const thumbUp =
    !fetching && !error && data?.alert?.feedback?.sentiment === 1 ? (
      <ThumbUp />
    ) : (
      <ThumbUpOutlined />
    )

  const isDown = data?.alert?.feedback?.sentiment === -1
  const thumbDown =
    !fetching && !error && isDown ? <ThumbDown /> : <ThumbDownOutlined />

  return (
    <Grid
      container
      spacing={1}
      justifyContent='flex-end'
      alignItems='center'
      sx={{ pr: '8px' }}
    >
      {mutationStatus.error?.message}
      <Grid item>
        <IconButton
          onClick={() => {
            setCacheCount(cacheCount + 1)
            commit(
              {
                input: {
                  alertID,
                  sentiment: 1,
                },
              },
              { additionalTypenames: ['Feedback'] },
            )
          }}
          size='large'
        >
          {thumbUp}
        </IconButton>
      </Grid>
      <Grid item>
        <IconButton
          onClick={() => {
            setCacheCount(cacheCount + 1)
            commit(
              {
                input: {
                  alertID,
                  sentiment: -1,
                },
              },
              { additionalTypenames: ['Feedback'] },
            )
          }}
          size='large'
        >
          {thumbDown}
        </IconButton>
      </Grid>
      <Grid item>
        <Tooltip title='Was this alert actionable?' placement='top'>
          <Info fontSize='small' sx={{ p: '12px' }} />
        </Tooltip>
      </Grid>
      <Grow in={isDown} mountOnEnter unmountOnExit>
        <Grid
          item
          xs={12}
          sx={{
            display: 'flex',
            justifyContent: 'flex-end',
          }}
        >
          <TextField
            className={classes.transition}
            placeholder='Why?'
            value={note}
            onChange={(e) => setNote(e.target.value)}
          />
        </Grid>
      </Grow>
    </Grid>
  )
}
