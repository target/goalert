import React from 'react'
import Card from '@mui/material/Card'
import Divider from '@mui/material/Divider'
import Grid from '@mui/material/Grid'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import ListItemSecondaryAction from '@mui/material/ListItemSecondaryAction'
import Typography from '@mui/material/Typography'
import { useTheme } from '@mui/material'
import TrueIcon from 'mdi-material-ui/CheckboxMarkedCircleOutline'
import FalseIcon from 'mdi-material-ui/CloseCircleOutline'
import FalseOkIcon from 'mdi-material-ui/MinusCircleOutline'
import { SWONode as SWONodeType } from '../../../schema'

interface SWONodeProps {
  node: SWONodeType
  name: string
}

export default function SWONode({ node, name }: SWONodeProps): JSX.Element {
  const theme = useTheme()

  return (
    <Grid item sx={{ minWidth: 350 }}>
      <Card raised={node.isLeader}>
        <Typography color={theme.palette.primary.main} sx={{ p: 2 }}>
          {name}
        </Typography>
        <List>
          <ListItem>
            <ListItemText primary='Executable?' />
            <ListItemSecondaryAction>
              {node.canExec ? (
                <TrueIcon color='success' />
              ) : (
                <FalseOkIcon color='disabled' />
              )}
            </ListItemSecondaryAction>
          </ListItem>
          <ListItem>
            <ListItemText primary='Old DB connection valid?' />
            <ListItemSecondaryAction>
              {node.oldValid ? (
                <TrueIcon color='success' />
              ) : (
                <FalseIcon color='error' />
              )}
            </ListItemSecondaryAction>
          </ListItem>
          <ListItem>
            <ListItemText primary='New DB connection valid?' />
            <ListItemSecondaryAction>
              {node.newValid ? (
                <TrueIcon color='success' />
              ) : (
                <FalseIcon color='error' />
              )}
            </ListItemSecondaryAction>
          </ListItem>
        </List>
      </Card>
    </Grid>
  )
}
