import React, { useState } from 'react'
import {
  ClickAwayListener,
  Divider,
  Drawer,
  Grid,
  List,
  ListItem,
  ListItemText,
  TextField,
  Toolbar,
  Typography,
  Button,
  ButtonGroup,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { GQLAPIKey, UpdateGQLAPIKeyInput } from '../../../schema'
import AdminAPIKeysDeleteDialog from './AdminAPIKeysDeleteDialog'
import { gql, useMutation } from '@apollo/client'
import { GenericError } from '../../error-pages'

const updateGQLAPIKeyQuery = gql`
  mutation UpdateGQLAPIKey($input: UpdateGQLAPIKeyInput!) {
    updateGQLAPIKey(input: $input)
  }
`

// const MaxDetailsLength = 6 * 1024 // 6KiB

interface Props {
  onClose: () => void
  apiKey: GQLAPIKey | null
}

const useStyles = makeStyles(() => ({
  buttons: {
    textAlign: 'right',
    width: '30vw',
    padding: '15px 10px',
  },
}))

export default function AdminAPIKeysDrawer(props: Props): JSX.Element {
  const { onClose, apiKey } = props
  const classes = useStyles()
  const isOpen = Boolean(apiKey)
  const [deleteDialog, onDeleteDialog] = useState(false)
  const [showEdit, setShowEdit] = useState(true)
  const [showSave, setShowSave] = useState(false)
  const handleDeleteConfirmation = (): void => {
    onDeleteDialog(!deleteDialog)
  }
  let comma = ''
  const allowFieldsStr = apiKey?.allowedFields.map((inp: string): string => {
    const inpComma = comma + inp
    comma = ', '

    return inpComma
  })
  const [key, setKey] = useState<UpdateGQLAPIKeyInput>({
    id: apiKey?.id ?? '',
    name: apiKey?.name,
    description: apiKey?.description,
  })
  const [updateAPIKey, updateAPIKeyStatus] = useMutation(updateGQLAPIKeyQuery, {
    onCompleted: (data) => {
      if (data.updateGQLAPIKey) {
        setShowSave(!showSave)
        setShowEdit(!showEdit)
      }
    },
  })
  const { loading, data, error } = updateAPIKeyStatus
  // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
  const handleSave = () => {
    updateAPIKey({
      variables: {
        input: key,
      },
    }).then((result) => {
      if (!result.errors) {
        return result
      }
    })
  }

  if (error) {
    return <GenericError error={error.message} />
  }

  if (loading && !data) {
    // return <Spinner />
  }

  return (
    <React.Fragment>
      {deleteDialog ? (
        <AdminAPIKeysDeleteDialog
          onClose={onDeleteDialog}
          apiKey={props.apiKey}
          close={deleteDialog}
        />
      ) : null}
      <ClickAwayListener onClickAway={onClose} mouseEvent='onMouseUp'>
        <Drawer
          anchor='right'
          open={isOpen}
          variant='persistent'
          data-cy='debug-message-details'
        >
          <Toolbar />
          <Grid style={{ width: '30vw' }}>
            <Typography variant='h6' style={{ margin: '16px' }}>
              API Key Details
            </Typography>
            <Divider />
            <List disablePadding>
              <ListItem divider>
                {showSave ? (
                  <TextField
                    required
                    id='standard-required'
                    label='Name'
                    defaultValue={apiKey?.name}
                    onChange={(e) => {
                      const keyTemp = key
                      keyTemp.name = e.target.value
                      setKey(keyTemp)
                    }}
                    variant='standard'
                    sx={{ width: '100%' }}
                  />
                ) : (
                  <ListItemText primary='Name' secondary={apiKey?.name} />
                )}
              </ListItem>
              <ListItem divider>
                {showSave ? (
                  <TextField
                    id='standard-multiline-static'
                    label='Description'
                    required
                    multiline
                    rows={4}
                    defaultValue={apiKey?.description}
                    onChange={(e) => {
                      const keyTemp = key
                      keyTemp.description = e.target.value
                      setKey(keyTemp)
                    }}
                    variant='standard'
                    sx={{ width: '100%' }}
                  />
                ) : (
                  <ListItemText
                    primary='Description'
                    secondary={apiKey?.description}
                  />
                )}
              </ListItem>
              <ListItem divider>
                <ListItemText
                  primary='Allowed Fields'
                  secondary={allowFieldsStr}
                />
              </ListItem>
              <ListItem divider>
                <ListItemText
                  primary='Creation Time'
                  secondary={apiKey?.createdAt}
                />
              </ListItem>
              <ListItem divider>
                <ListItemText
                  primary='Created By'
                  secondary={apiKey?.createdBy?.name}
                />
              </ListItem>
              <ListItem divider>
                <ListItemText
                  primary='Expires At'
                  secondary={apiKey?.expiresAt}
                />
              </ListItem>
              <ListItem divider>
                <ListItemText
                  primary='Updated By'
                  secondary={apiKey?.updatedBy?.name}
                />
              </ListItem>
            </List>
            <Grid className={classes.buttons}>
              {showSave ? (
                <ButtonGroup variant='outlined'>
                  <Button
                    data-cy='exit'
                    onClick={() => {
                      setShowSave(!showSave)
                      setShowEdit(!showEdit)
                    }}
                  >
                    EXIT
                  </Button>
                  <Button
                    data-cy='save'
                    // disabled={isEmpty(apiKey?.description + apiKey?.name)}
                    onClick={handleSave}
                  >
                    SAVE
                  </Button>
                </ButtonGroup>
              ) : null}
              {showEdit ? (
                <ButtonGroup variant='outlined'>
                  <Button data-cy='delete' onClick={handleDeleteConfirmation}>
                    DELETE
                  </Button>
                  <Button
                    data-cy='edit'
                    onClick={() => {
                      setShowSave(!showSave)
                      setShowEdit(!showEdit)
                    }}
                  >
                    EDIT
                  </Button>
                </ButtonGroup>
              ) : null}
            </Grid>
          </Grid>
        </Drawer>
      </ClickAwayListener>
    </React.Fragment>
  )
}
