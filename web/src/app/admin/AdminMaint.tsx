import React from 'react'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Button,
  Card,
  CardActions,
  CardContent,
  CardHeader,
} from '@mui/material'
import Markdown from '../util/Markdown'
import KeyRotation from './KeyRotation.md'
import { ExpandMore } from '@mui/icons-material'
import FormDialog from '../dialogs/FormDialog'
import { gql, useMutation } from 'urql'
import { useErrorConsumer } from '../util/ErrorConsumer'

type AdminMaintToolProps = {
  name: string
  desc: string
  guide: string
  children: React.ReactNode
}

function AdminMaintTool(props: AdminMaintToolProps): React.ReactNode {
  return (
    <Grid item xs={6}>
      <Card>
        <CardHeader title={props.name} />
        <CardContent>
          <Typography>{props.desc}</Typography>
          {props.guide && (
            <Accordion>
              <AccordionSummary
                expandIcon={<ExpandMore />}
                id='panel-header'
                aria-controls='panel-content'
              >
                Guide
              </AccordionSummary>
              <AccordionDetails>
                <Markdown value={props.guide} />
              </AccordionDetails>
            </Accordion>
          )}
        </CardContent>
        <CardActions>{props.children}</CardActions>
      </Card>
    </Grid>
  )
}

export default function AdminMaint(): React.JSX.Element {
  const [reEncryptOpen, setReEncryptOpen] = React.useState(false)
  return (
    <Grid container spacing={2}>
      <AdminMaintTool
        name='Encryption Key Rotation'
        desc='Keyrings and configuration are encryped at-rest, you may wish to rotate the encryption key periodically.'
        guide={KeyRotation}
      >
        <Button
          size='small'
          onClick={() => {
            setReEncryptOpen(true)
          }}
        >
          Re-Encrypt Data
        </Button>
        {reEncryptOpen && (
          <ReEncryptData onClose={() => setReEncryptOpen(false)} />
        )}
      </AdminMaintTool>
    </Grid>
  )
}

const reEncryptMut = gql`
  mutation {
    reEncryptKeyringsAndConfig
  }
`

type ReEncryptDataProps = {
  onClose: () => void
}

function ReEncryptData(props: ReEncryptDataProps): React.ReactNode {
  const [status, commit] = useMutation(reEncryptMut)
  const errs = useErrorConsumer(status.error)

  return (
    <FormDialog
      title='Re-Encrypt all data?'
      subTitle='This will re-encrypt all keyrings and configuration data with the latest encryption key and is not reversible.'
      onClose={props.onClose}
      onSubmit={() =>
        commit().then((res) => {
          if (res.error) return

          props.onClose()
        })
      }
      primaryActionLabel='Confirm'
      loading={status.fetching}
      errors={errs.remainingLegacy()}
    />
  )
}
