import { gql, useQuery, useMutation } from '@apollo/client'
import React, { useState } from 'react'
import p from 'prop-types'

import { SetFavoriteButton } from './SetFavoriteButton'
import { oneOfShape } from '../util/propTypes'
import { Button, Dialog, DialogActions, DialogTitle } from '@material-ui/core'
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
}

const mutation = gql`
  mutation setFav($input: SetFavoriteInput!) {
    setFavorite(input: $input)
  }
`

export function QuerySetFavoriteButton(props) {
  let id, typeName
  if (props.rotationID) {
    typeName = 'rotation'
    id = props.rotationID
  } else if (props.serviceID) {
    typeName = 'service'
    id = props.serviceID
  } else if (props.scheduleID) {
    typeName = 'schedule'
    id = props.scheduleID
  } else {
    throw new Error('unknown type')
  }
  const { data, loading } = useQuery(queries[typeName], {
    variables: { id },
  })
  const isFavorite = data && data.data && data.data.isFavorite
  const [showMutationErrorDialog, setShowMutationErrorDialog] = useState(false)
  const [toggleFav, toggleFavStatus] = useMutation(mutation, {
    variables: {
      input: { target: { id, type: typeName }, favorite: !isFavorite },
    },
    onError: () => setShowMutationErrorDialog(true),
  })

  return (
    <React.Fragment>
      <SetFavoriteButton
        typeName={typeName}
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
            color='primary'
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

QuerySetFavoriteButton.propTypes = {
  id: oneOfShape({
    serviceID: p.string,
    rotationID: p.string,
    scheduleID: p.string,
  }),
}
