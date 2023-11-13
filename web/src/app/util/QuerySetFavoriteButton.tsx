import React, { useState } from 'react'
import { gql, useQuery, useMutation } from '@apollo/client'
import { SetFavoriteButton } from './SetFavoriteButton'
import { Button, Dialog, DialogActions, DialogTitle } from '@mui/material'
import DialogContentError from '../dialogs/components/DialogContentError'

const queries = {
  service: gql`
    query serviceFavQuery($id: ID!) {
      data: service(id: $id) {
        id
        isFavorite
      }
    }
  `,
  rotation: gql`
    query rotationFavQuery($id: ID!) {
      data: rotation(id: $id) {
        id
        isFavorite
      }
    }
  `,
  schedule: gql`
    query scheduleFavQuery($id: ID!) {
      data: schedule(id: $id) {
        id
        isFavorite
      }
    }
  `,
  escalationPolicy: gql`
    query escalationPolicyFavQuery($id: ID!) {
      data: escalationPolicy(id: $id) {
        id
        isFavorite
      }
    }
  `,
  user: gql`
    query userFavQuery($id: ID!) {
      data: user(id: $id) {
        id
        isFavorite
      }
    }
  `,
}

const mutation = gql`
  mutation setFav($input: SetFavoriteInput!) {
    setFavorite(input: $input)
  }
`

interface QuerySetFavoriteButtonProps {
  id: string
  type: 'rotation' | 'service' | 'schedule' | 'escalationPolicy' | 'user'
}

export function QuerySetFavoriteButton({
  id,
  type,
}: QuerySetFavoriteButtonProps): React.ReactNode {
  const { data, loading } = useQuery(queries[type], {
    variables: { id },
  })
  const isFavorite = data && data.data && data.data.isFavorite
  const [showMutationErrorDialog, setShowMutationErrorDialog] = useState(false)
  const [toggleFav, toggleFavStatus] = useMutation(mutation, {
    variables: {
      input: { target: { id, type }, favorite: !isFavorite },
    },
    onError: () => setShowMutationErrorDialog(true),
  })

  return (
    <React.Fragment>
      <SetFavoriteButton
        typeName={type}
        isFavorite={isFavorite}
        loading={!data && loading}
        onClick={() => toggleFav()}
      />
      <Dialog
        // if showMutationErrorDialog is true, the dialog will open
        open={showMutationErrorDialog}
        // onClose, reset showMutationErrorDialog to false
        onClose={() => setShowMutationErrorDialog(false)}
      >
        <DialogTitle>An error occurred</DialogTitle>
        <DialogContentError error={toggleFavStatus?.error?.message ?? ''} />
        <DialogActions>
          <Button
            variant='contained'
            onClick={() => setShowMutationErrorDialog(false)}
          >
            Okay
          </Button>
        </DialogActions>
      </Dialog>
    </React.Fragment>
  )
}
