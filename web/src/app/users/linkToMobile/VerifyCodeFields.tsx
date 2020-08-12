import React, { useState } from 'react'
import {
  DialogContent,
  DialogContentText,
  Grid,
  TextField,
  makeStyles,
} from '@material-ui/core'

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

export default function VerifyCodeFields() {
  const classes = useStyles()
  const [numOne, setNumOne] = useState('')
  const [numTwo, setNumTwo] = useState('')
  const [numThree, setNumThree] = useState('')
  const [numFour, setNumFour] = useState('')

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
                if (numOne && numTwo && numThree) {
                  // todo: go to next page and submit claim code automatically
                }
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
