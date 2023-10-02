import React, { useState } from 'react'
import {
  ClickAwayListener,
  Divider,
  Drawer,
  Grid,
  List,
  ListItem,
  ListItemText,
  Toolbar,
  Typography,
  Button,
  ButtonGroup,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { GQLAPIKey } from '../../../schema'
import AdminAPIKeysDeleteDialog from './AdminAPIKeysDeleteDialog'
import { Time } from '../../util/Time'

// property for this object
interface Props {
  onClose: () => void
  apiKey: GQLAPIKey
  setCreate: (param: boolean) => void
  setAPIKey: (param: GQLAPIKey) => void
  setOpenActionAPIKeyDialog: (param: boolean) => void
}

const useStyles = makeStyles(() => ({
  buttons: {
    textAlign: 'right',
    width: '30vw',
    padding: '15px 10px',
  },
}))

export default function AdminAPIKeysDrawer(props: Props): JSX.Element {
  const { onClose, apiKey, setCreate, setOpenActionAPIKeyDialog, setAPIKey } =
    props
  const classes = useStyles()
  const isOpen = Boolean(apiKey)
  const [deleteDialog, onDeleteDialog] = useState(false)
  // handle for opening/closing delete confirmation dialog of the API Key Delete transaction
  const handleDeleteConfirmation = (): void => {
    onDeleteDialog(!deleteDialog)
  }
  let comma = ''
  // convert allowedfields option array data to comma separated values which will be use for display
  const allowFieldsStr = apiKey?.allowedFields.map((inp: string): string => {
    const inpComma = comma + inp
    comma = ', '
    return inpComma
  })

  return (
    <React.Fragment>
      {deleteDialog ? (
        <AdminAPIKeysDeleteDialog
          onClose={onDeleteDialog}
          apiKey={props.apiKey}
          close={deleteDialog}
        />
      ) : null}
      <ClickAwayListener onClickAway={onClose}>
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
                <ListItemText primary='Name' secondary={apiKey?.name} />
              </ListItem>
              <ListItem divider>
                <ListItemText
                  primary='Description'
                  secondary={apiKey?.description}
                />
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
                  secondary={<Time prefix='' time={apiKey?.createdAt} />}
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
                  secondary={<Time prefix='' time={apiKey?.expiresAt} />}
                />
              </ListItem>
              <ListItem divider>
                <ListItemText
                  primary='Updated By'
                  secondary={apiKey?.updatedBy?.name}
                />
              </ListItem>
              <ListItem divider>
                <ListItemText primary='Role' secondary={apiKey?.role} />
              </ListItem>
            </List>
            <Grid className={classes.buttons}>
              <ButtonGroup variant='contained'>
                <Button data-cy='delete' onClick={handleDeleteConfirmation}>
                  DELETE
                </Button>
                <Button
                  data-cy='edit'
                  onClick={() => {
                    setAPIKey(apiKey)
                    setCreate(false)
                    setOpenActionAPIKeyDialog(true)
                  }}
                >
                  EDIT
                </Button>
              </ButtonGroup>
            </Grid>
          </Grid>
        </Drawer>
      </ClickAwayListener>
    </React.Fragment>
  )
}
