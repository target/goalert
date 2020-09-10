import React, { useEffect, useState } from 'react'
import {
  Button,
  DialogContent,
  DialogContentText,
  Grid,
  TextField,
  makeStyles,
} from '@material-ui/core'
import gql from 'graphql-tag'
import { useMutation } from 'react-apollo'
import { useDispatch } from 'react-redux'
import { setURLParam } from '../../actions'
import { getAllyColors, Color } from '../../util/colors'

const useStyles = makeStyles({
  buttonContainer: {
    display: 'flex',
    justifyContent: 'center',
  },
  codeContainer: {
    margin: '0.5em',
  },
  contentText: {
    marginBottom: 0,
  },
  textField: {
    textAlign: 'center',
    fontSize: '2.25rem',
  },
})

const mutation = gql`
  mutation($input: VerifyAuthLinkInput!) {
    verifyAuthLink(input: $input)
  }
`

interface VerifyCodeFieldsProps {
  authLinkID: string
  verifyCode: string
}

export default function VerifyCodeFields(
  props: VerifyCodeFieldsProps,
): JSX.Element {
  const classes = useStyles()
  const [colors, setColors] = useState(null as Color[] | null)

  const dispatch = useDispatch()
  const setErrorMessage = (value: string): void => {
    dispatch(setURLParam('error', value))
  }

  useEffect(() => {
    setColors(getAllyColors(props.verifyCode, 4))
  }, [])

  const [verifyCode] = useMutation(mutation, {
    variables: {
      input: {
        id: props.authLinkID,
        code: props.verifyCode,
      },
    },

    onError: (err) => {
      if (err.message) setErrorMessage(err.message)
    },
  })

  const renderTextField = (i: number): JSX.Element => (
    <Grid item xs={3}>
      <TextField
        value={props.verifyCode.charAt(i)}
        InputProps={{
          readOnly: true,
        }}
        inputProps={{
          className: classes.textField,
          style: {
            color:
              colors && colors[i]
                ? `rgb(${colors[i][0]}, ${colors[i][1]}, ${colors[i][2]})`
                : 'rgb(0, 0, 0)',
          },
        }}
      />
    </Grid>
  )

  return (
    <DialogContent>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <DialogContentText className={classes.contentText}>
            Please verify that the code displayed is the same on your mobile
            device.
          </DialogContentText>
        </Grid>
        <Grid
          className={classes.codeContainer}
          item
          xs={12}
          container
          spacing={2}
        >
          {renderTextField(0)}
          {renderTextField(1)}
          {renderTextField(2)}
          {renderTextField(3)}
        </Grid>
        <Grid className={classes.buttonContainer} item xs={12}>
          <Button
            variant='contained'
            color='primary'
            onClick={() => verifyCode()}
          >
            Looks good
          </Button>
        </Grid>
      </Grid>
    </DialogContent>
  )
}
