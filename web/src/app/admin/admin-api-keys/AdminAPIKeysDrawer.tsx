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

  const handleDeleteConfirmation = (): void => {
    onDeleteDialog(!deleteDialog)
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
                  secondary={apiKey?.allowedFields}
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
              <ButtonGroup variant='outlined'>
                <Button
                  data-cy='delete'
                  // disabled={isEmpty(values)}
                  onClick={handleDeleteConfirmation}
                >
                  DELETE
                </Button>
                <Button
                  data-cy='save'
                  // disabled={isEmpty(values)}
                  // onClick={() => setValues({})}
                >
                  SAVE
                </Button>
              </ButtonGroup>
            </Grid>
          </Grid>
        </Drawer>
      </ClickAwayListener>
    </React.Fragment>
  )
}
