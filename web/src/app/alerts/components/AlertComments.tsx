import React, { useState } from 'react'
import { gql, useQuery, useMutation } from 'urql'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'
import FormControlLabel from '@mui/material/FormControlLabel'
import Switch from '@mui/material/Switch'
import { Time } from '../../util/Time'
import Markdown from '../../util/Markdown'
import { UserAvatar } from '../../util/avatars'
import ListItemAvatar from '@mui/material/ListItemAvatar'
import Avatar from '@mui/material/Avatar'
import PersonIcon from '@mui/icons-material/Person'

const MAX_COMMENT_LENGTH = 4096

const query = gql`
  query AlertComments($id: Int!) {
    alert(id: $id) {
      id
      comments {
        id
        body
        createdAt
        user {
          id
          name
        }
      }
    }
  }
`

const addMutation = gql`
  mutation AddAlertComment($input: AddAlertCommentInput!) {
    addAlertComment(input: $input) {
      id
    }
  }
`

interface AlertCommentsProps {
  alertID: number

  /*
   * Shared with the Event Log rather than held locally, so "Full Timestamps"
   * is one preference for the page -- toggling it on either card updates both,
   * and it persists through the same localStorage key.
   */
  showExactTimes?: boolean
  onToggleExactTimes?: () => void
}

export default function AlertComments(
  props: AlertCommentsProps,
): React.JSX.Element {
  const { alertID, showExactTimes, onToggleExactTimes } = props
  const [body, setBody] = useState('')

  const [{ data, fetching: loading }] = useQuery({
    query,
    variables: { id: alertID },
    // Without this, adding the *first* comment does not refresh the list:
    // urql's document cache keys off typenames present in the response, and an
    // empty comments array contains no AlertComment for the mutation to
    // invalidate against.
    context: React.useMemo(
      () => ({ additionalTypenames: ['AlertComment'] }),
      [],
    ),
  })
  const [addStatus, add] = useMutation(addMutation)

  const comments = data?.alert?.comments ?? []
  const trimmed = body.trim()
  const tooLong = body.length > MAX_COMMENT_LENGTH

  function submit(): void {
    if (!trimmed || tooLong) return
    add({ input: { alertID, body: trimmed } }).then((res) => {
      if (res.error) return
      setBody('')
    })
  }

  // Returns a bare Card -- the caller owns grid placement. Returning a Grid
  // item from here nested one inside another and collapsed the card to a
  // narrow column.
  return (
    <Card style={{ width: '100%' }}>
      {/* Header laid out to match the Event Log card. */}
      <div style={{ display: 'flex' }}>
        <CardContent style={{ flex: 1, paddingBottom: 0 }}>
          <Typography component='h3' variant='h5'>
            Comments
          </Typography>
        </CardContent>
        {onToggleExactTimes && (
          <FormControlLabel
            control={
              <Switch
                checked={Boolean(showExactTimes)}
                onChange={onToggleExactTimes}
                data-cy='toggle-comment-times'
              />
            }
            label='Full Timestamps'
            style={{ padding: '0.5em 0.5em 0 0' }}
          />
        )}
      </div>
      <CardContent style={{ paddingTop: 0 }}>
        {comments.length === 0 && !loading && (
          <Typography color='textSecondary' variant='body2' sx={{ pt: 1 }}>
            No comments yet.
          </Typography>
        )}

        <List data-cy='alert-comments'>
          {comments.map(
            (c: {
              id: string
              body: string
              createdAt: string
              user?: { id: string; name: string } | null
            }) => (
              <ListItem key={c.id} alignItems='flex-start' disableGutters>
                <ListItemAvatar>
                  {c.user ? (
                    <UserAvatar userID={c.user.id} />
                  ) : (
                    <Avatar>
                      <PersonIcon />
                    </Avatar>
                  )}
                </ListItemAvatar>
                <ListItemText
                  // Markdown emits block elements, so the secondary slot
                  // must not default to a <p>.
                  secondaryTypographyProps={{ component: 'div' }}
                  primary={
                    <React.Fragment>
                      {/* A comment outlives its author, so the account may
                            be gone -- say so rather than rendering a blank. */}
                      <strong>{c.user?.name ?? 'Deleted user'}</strong>
                      <Typography
                        component='span'
                        variant='body2'
                        color='textSecondary'
                        sx={{ pl: 1 }}
                      >
                        <Time
                          time={c.createdAt}
                          format={showExactTimes ? 'default' : 'relative'}
                        />
                      </Typography>
                    </React.Fragment>
                  }
                  secondary={<Markdown value={c.body} />}
                />
              </ListItem>
            ),
          )}
        </List>

        <TextField
          fullWidth
          multiline
          minRows={2}
          placeholder='Add a comment'
          name='comment-body'
          data-cy='new-comment'
          value={body}
          error={tooLong}
          helperText={
            tooLong
              ? `Comments cannot exceed ${MAX_COMMENT_LENGTH} characters.`
              : undefined
          }
          onChange={(e) => setBody(e.target.value)}
        />

        {addStatus.error && (
          <Typography color='error' variant='body2' sx={{ pt: 1 }}>
            {addStatus.error.message}
          </Typography>
        )}

        <Button
          sx={{ mt: 1 }}
          variant='contained'
          aria-label='Add Comment'
          disabled={!trimmed || tooLong || addStatus.fetching}
          onClick={submit}
        >
          Comment
        </Button>
      </CardContent>
    </Card>
  )
}
