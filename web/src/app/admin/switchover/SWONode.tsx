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
import { SWONode as SWONodeType } from '../../../schema'

interface SWONodeProps {
  node: SWONodeType
  index: number
}

export default function SWONode({ node, index }: SWONodeProps): JSX.Element {
  const theme = useTheme()

  return (
    <Grid item key={index} sx={{ minWidth: 300 }}>
      <Card>
        <Typography color={theme.palette.primary.main} sx={{ p: 2 }}>
          Node {index + 1}
        </Typography>
        <List
          subheader={
            <React.Fragment>
              <Divider />
              <Typography color={theme.palette.secondary.main} sx={{ p: 2 }}>
                Status: {node.status}
              </Typography>
              <Divider />
            </React.Fragment>
          }
        >
          <ListItem>
            <ListItemText primary='Executable?' />
            <ListItemSecondaryAction>
              {node.canExec ? (
                <TrueIcon color='success' />
              ) : (
                <FalseIcon color='error' />
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
