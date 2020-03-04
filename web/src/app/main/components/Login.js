import React, { Component } from 'react'
import axios from 'axios'
import Button from '@material-ui/core/Button'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import Typography from '@material-ui/core/Typography'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import Divider from '@material-ui/core/Divider'
import Hidden from '@material-ui/core/Hidden'
import isFullScreen from '@material-ui/core/withMobileDialog'
import { withStyles } from '@material-ui/core/styles'
import { getParameterByName } from '../../util/query_param'

import logoSrcSet1 from '../../public/goalert-logo-scaled.webp'
import logoSrcSet2 from '../../public/goalert-logo-scaled@1.5.webp'
import logoSrcSet3 from '../../public/goalert-logo-scaled@2.webp'
import logoImgSrc from '../../public/goalert-logo-scaled@2.png'
import { pathPrefix } from '../../env'

const PROVIDERS = pathPrefix + '/api/v2/identity/providers'
const BACKGROUND_URL =
  'https://www.toptal.com/designers/subtlepatterns/patterns/dust_scratches.png'

const styles = {
  card: {
    width: 'fit-content',
    maxWidth: '30em',
  },
  center: {
    position: 'fixed',
    top: '40%',
    left: '50%',
    WebkitTransform: 'translate(-50%, -40%)',
    transform: 'translateY(-50%, -40%)',
    textAlign: 'center',
  },
  divider: {
    width: '9em',
  },
  error: {
    color: 'red',
  },
  footer: {
    paddingBottom: '0.5em',
  },
  gridContainer: {
    width: 'fit-content',
  },
  hasNext: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  loginIcon: {
    height: '1.5em',
    width: '1.5em',
    paddingRight: '0.5em',
  },
  or: {
    paddingLeft: '1em',
    paddingRight: '1em',
  },
}

@withStyles(styles)
@isFullScreen()
export default class Login extends Component {
  constructor(props) {
    super(props)

    this.state = {
      error: getParameterByName('login_error') || '',
      providers: [],
    }
  }

  componentWillReceiveProps(next) {
    if (this.props.fullScreen === next.fullScreen) {
      return
    }

    this.setBackground(next.fullScreen)
  }

  componentDidMount() {
    this.setBackground(this.props.fullScreen)

    // get providers
    axios
      .get(PROVIDERS)
      .then(res => this.setState({ providers: res.data }))
      .catch(err => this.setState({ error: err }))
  }

  /*
   * Sets the background image for the login page
   *
   * Background pattern from Toptal Subtle Patterns
   */
  setBackground = fullScreen => {
    if (fullScreen) {
      document.body.style.backgroundColor = `white` // overrides light grey background
    } else {
      document.body.style.backgroundImage = `url('${BACKGROUND_URL}')` // overrides light grey background
    }
  }

  /*
   * Renders a field from a provider
   */
  renderField = field => {
    const {
      ID: id, // unique name/identifier of the field
      Label: label, // placeholder text that is displayed to the use in the field
      Password: password, // indicates that a field should be treated as a password
      Required: required, // indicates that a field must not be empty
      // Scannable: scannable todo: indicates that the field can be entered via QR-code scan
    } = field

    return (
      <Grid key={id} item xs={12}>
        <TextField
          label={label}
          required={required}
          type={password ? 'password' : 'text'}
          name={id}
        />
      </Grid>
    )
  }

  /*
   * Renders a divider if there is another provider after
   */
  renderHasNextDivider = (idx, len) => {
    const { classes } = this.props

    if (idx + 1 < len) {
      return (
        <Grid item xs={12} className={classes.hasNext}>
          <Divider className={classes.divider} />
          <Typography className={classes.or}>or</Typography>
          <Divider className={classes.divider} />
        </Grid>
      )
    }
  }

  /*
   * Renders a provider given from initial GET request
   */
  renderProvider = (provider, idx, len) => {
    const { classes } = this.props
    const {
      ID: id, // unique identifier of the provider
      Fields: fields, // holds a list of fields to include with the request
      Hidden: hidden, // indicates that the provider is not intended for user visibility
      LogoUrl: logoUrl, // optional URL of an icon to display with the provider
      Title: title, // user-viable string for identifying this provider
      URL: url, // the location of the form action (POST)
    } = provider

    if (hidden) return

    // create login button
    let loginButton = null
    const loginIcon = logoUrl ? (
      <img alt='GoAlert' src={logoUrl} className={classes.loginIcon} />
    ) : null
    if (fields) {
      loginButton = (
        <Button type='submit' variant='contained' color='primary'>
          Login
        </Button>
      )
    } else {
      loginButton = (
        <Button type='submit' variant='contained' color='primary'>
          {loginIcon}
          Login with {title}
        </Button>
      )
    }

    let form = null
    if (fields && fields.length) {
      form = (
        <Grid container spacing={2}>
          {fields.map(field => this.renderField(field))}
          <Grid item xs={12}>
            {loginButton}
          </Grid>
        </Grid>
      )
    } else {
      form = loginButton
    }

    return (
      <React.Fragment key={idx}>
        <Grid item xs={12}>
          <form action={url} method='post' id={'auth-' + id}>
            {form}
          </form>
        </Grid>
        {this.renderHasNextDivider(idx, len)}
      </React.Fragment>
    )
  }

  render() {
    const { classes } = this.props
    const { error, providers } = this.state

    // error message if GET fails
    let errorJSX = null
    if (error) {
      errorJSX = (
        <Grid item xs={12}>
          <Typography variant='subtitle1' className={classes.error}>
            {error.toString()}
          </Typography>
        </Grid>
      )
    }

    const logo = (
      <picture>
        <source
          srcSet={`
              ${logoSrcSet1},
              ${logoSrcSet2} 1.5x,
              ${logoSrcSet3} 2x
            `}
          type='image/webp'
        />
        <img src={logoImgSrc} height={61} alt='GoAlert' />
      </picture>
    )

    return (
      <React.Fragment>
        <Hidden smDown>
          <div className={classes.center}>
            <Card className={classes.card}>
              <CardContent>
                <Grid container spacing={2} className={classes.gridContainer}>
                  <Grid item xs={12}>
                    {logo}
                  </Grid>
                  {providers.map((provider, idx) =>
                    this.renderProvider(provider, idx, providers.length),
                  )}
                  {errorJSX}
                </Grid>
              </CardContent>
            </Card>
          </div>
        </Hidden>
        <Hidden mdUp>
          <div className={classes.center}>
            <Grid container spacing={2} className={classes.gridContainer}>
              <Grid item xs={12}>
                {logo}
              </Grid>
              {providers.map((provider, idx) =>
                this.renderProvider(provider, idx, providers.length),
              )}
              {errorJSX}
            </Grid>
          </div>
        </Hidden>
      </React.Fragment>
    )
  }
}
