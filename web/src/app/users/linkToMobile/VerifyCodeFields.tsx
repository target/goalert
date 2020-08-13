import React, { useEffect, useState } from 'react'
import {
  DialogContent,
  DialogContentText,
  Grid,
  TextField,
  makeStyles,
} from '@material-ui/core'
import gql from 'graphql-tag'
import { useMutation } from 'react-apollo'

const useStyles = makeStyles({
  codeContainer: {
    marginTop: '2em',
  },
  contentText: {
    marginBottom: 0,
  },
  textField: {
    textAlign: 'center',
    fontSize: '1.5rem',
  },
})

const mutation = gql`
  mutation($input: VerifyAuthLinkInput!) {
    verifyAuthLink(input: $input)
  }
`

interface VerifyCodeFieldsProps {
  authLinkID: string
}

export default function VerifyCodeFields(props: VerifyCodeFieldsProps) {
  const classes = useStyles()
  const [numOne, setNumOne] = useState('')
  const [numTwo, setNumTwo] = useState('')
  const [numThree, setNumThree] = useState('')
  const [numFour, setNumFour] = useState('')

  const [verifyCode, verifyCodeStatus] = useMutation(mutation, {
    variables: {
      input: {
        id: props.authLinkID,
        code: numOne + numTwo + numThree + numFour,
      },
    },
  })

  useEffect(() => {
    if (numOne && numTwo && numThree && numFour) {
      verifyCode()
    }
  }, [numOne, numTwo, numThree, numFour])

  return (
    <DialogContent>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <DialogContentText className={classes.contentText}>
            Enter the code displayed on your mobile device.
          </DialogContentText>
        </Grid>
        <Grid
          className={classes.codeContainer}
          item
          xs={12}
          container
          spacing={2}
        >
          <Grid item xs={3}>
            <TextField
              id='numOne'
              value={numOne}
              onChange={(e) => {
                setNumOne(e.target.value)
                if (!numTwo) {
                  document.getElementById('numTwo')?.focus()
                }
              }}
              inputProps={{
                maxLength: 1,
                className: classes.textField,
              }}
            />
          </Grid>
          <Grid item xs={3}>
            <TextField
              id='numTwo'
              value={numTwo}
              onChange={(e) => {
                setNumTwo(e.target.value)
                if (!numThree) {
                  document.getElementById('numThree')?.focus()
                }
              }}
              inputProps={{
                maxLength: 1,
                className: classes.textField,
              }}
            />
          </Grid>
          <Grid item xs={3}>
            <TextField
              id='numThree'
              value={numThree}
              onChange={(e) => {
                setNumThree(e.target.value)
                if (!numFour) {
                  document.getElementById('numFour')?.focus()
                }
              }}
              inputProps={{
                maxLength: 1,
                className: classes.textField,
              }}
            />
          </Grid>
          <Grid item xs={3}>
            <TextField
              id='numFour'
              value={numFour}
              onChange={(e) => {
                setNumFour(e.target.value)
              }}
              inputProps={{
                maxLength: 1,
                className: classes.textField,
              }}
            />
          </Grid>
        </Grid>
      </Grid>
    </DialogContent>
  )
}
