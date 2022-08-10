import React, { useState } from 'react'
import {
  Button,
  Card,
  CardActions,
  CardContent,
  CardHeader,
  Grid,
  Typography,
} from '@mui/material'
import { OpenInNew } from 'mdi-material-ui'
import { ConfigID } from '../../schema'
import AdminDialog from '../admin/AdminDialog'

export default function LocalDev() {
  const [updateConfig, setUpdateConfig] = useState<Partial<
    Record<ConfigID, string>
  > | null>(null)

  return (
    <React.Fragment>
      <Grid container>
        <Card>
          <CardHeader title='SMTP' />
          <CardContent>
            <Typography>
              Allow email contact methods and a local mail server.
            </Typography>
          </CardContent>
          <CardActions>
            <Button
              title='Update GoAlert config to use the dev mail server.'
              size='small'
              onClick={() => {
                setUpdateConfig({
                  'SMTP.Enable': 'true',
                  'SMTP.From': 'goalert@localhost',
                  'SMTP.Address': 'localhost:1025',
                  'SMTP.Username': '',
                  'SMTP.Password': '',
                  'SMTP.DisableTLS': 'true',
                })
              }}
            >
              Update Config
            </Button>
            <Button target='_blank' size='small' href='http://localhost:8025'>
              View Messages <OpenInNew fontSize='small' />
            </Button>
          </CardActions>
        </Card>
      </Grid>

      {updateConfig && (
        <AdminDialog
          value={updateConfig}
          onClose={() => setUpdateConfig(null)}
        />
      )}
    </React.Fragment>
  )
}
