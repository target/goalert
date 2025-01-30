import React from 'react'
import Card from '@mui/material/Card'
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

export default function SWONode({
  node,
  name,
}: SWONodeProps): React.JSX.Element {
  const theme = useTheme()

  if (node.id.startsWith('unknown-')) {
    return (
      <Grid item xs={12} sm={6} lg={4} xl={3} sx={{ width: '100%' }}>
        <Card>
          <Typography color={theme.palette.primary.main} sx={{ p: 2 }}>
            {name}
          </Typography>
          <List>
            <ListItem>
              <ListItemText
                primary='Application'
                secondary={node.id.substring(8) || '(No name given)'}
              />

              {node.id.includes('GoAlert') && (
                <ListItemSecondaryAction title='All GoAlert instances must be in switchover mode or dataloss could occur.'>
                  <FalseIcon color='error' />
                </ListItemSecondaryAction>
              )}
            </ListItem>
            <ListItem>
              <ListItemText primary='Connections' />
              <ListItemSecondaryAction
                sx={{ width: '24px', textAlign: 'center' }}
              >
                <Typography>
                  {node.connections?.reduce((acc, cur) => acc + cur.count, 0)}
                </Typography>
              </ListItemSecondaryAction>
            </ListItem>
          </List>
        </Card>
      </Grid>
    )
  }

  return (
    <Grid item xs={12} sm={6} lg={4} xl={3} sx={{ width: '100%' }}>
      <Card raised={node.isLeader}>
        <Typography color={theme.palette.primary.main} sx={{ p: 2 }}>
          {name}
        </Typography>
        <List>
          <ListItem>
            <ListItemText primary='Executable?' />
            <ListItemSecondaryAction title='Indicates a node capable of performing data replication.'>
              {node.canExec ? (
                <TrueIcon color='success' />
              ) : (
                <FalseOkIcon color='disabled' />
              )}
            </ListItemSecondaryAction>
          </ListItem>
          <ListItem>
            <ListItemText primary='Config valid?' />
            <ListItemSecondaryAction title={node.configError}>
              {!node.configError ? (
                <TrueIcon color='success' />
              ) : (
                <FalseIcon color='error' />
              )}
            </ListItemSecondaryAction>
          </ListItem>
          <ListItem>
            <ListItemText primary='Uptime' />
            <ListItemSecondaryAction
              title={
                node.uptime ? '' : 'Node appeared outside of reset window.'
              }
            >
              {node.uptime ? node.uptime : <FalseIcon color='error' />}
            </ListItemSecondaryAction>
          </ListItem>
          <ListItem>
            <ListItemText primary='Connections' />
            <ListItemSecondaryAction
              sx={{ width: '24px', textAlign: 'center' }}
            >
              <Typography>
                {node.connections?.reduce((acc, cur) => acc + cur.count, 0)}
              </Typography>
            </ListItemSecondaryAction>
          </ListItem>
        </List>
      </Card>
    </Grid>
  )
}
