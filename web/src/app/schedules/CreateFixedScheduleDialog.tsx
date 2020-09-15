import React from 'react'
import {
  AppBar,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  IconButton,
  Slide,
  SlideProps,
  Stepper,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Toolbar,
  Typography,
  makeStyles,
} from '@material-ui/core'
import CloseIcon from '@material-ui/icons/Close'

interface CreateFixedScheduleDialogProps {
  open: boolean
  onClose: () => void
}

const useStyles = makeStyles((theme) => ({
  appBar: {
    position: 'relative',
  },
  title: {
    marginLeft: theme.spacing(2),
    flex: 1,
  },
}))

const Transition = React.forwardRef<unknown, SlideProps>((props, ref) => (
  <Slide direction='up' ref={ref} {...props} />
))

export default function CreateFixedScheduleDialog({
  open,
  onClose,
}: CreateFixedScheduleDialogProps) {
  const classes = useStyles()

  function handleSubmit(): any {
    console.log('submitting')
    return null
  }

  return (
    <Dialog
      fullScreen
      open={open}
      onClose={onClose}
      TransitionComponent={Transition}
    >
      <AppBar className={classes.appBar}>
        <Toolbar>
          <IconButton
            edge='start'
            color='inherit'
            onClick={onClose}
            aria-label='close'
          >
            <CloseIcon />
          </IconButton>
          <Typography variant='h6' className={classes.title}>
            Created Fixed Schedule Adjustments
          </Typography>
          <Button autoFocus color='inherit' onClick={handleSubmit}>
            Submit
          </Button>
        </Toolbar>
      </AppBar>
    </Dialog>
  )
}
