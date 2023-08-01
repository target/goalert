import React, { useState, useEffect } from 'react'
import Button from '@mui/material/Button'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import Typography from '@mui/material/Typography'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import Divider from '@mui/material/Divider'
import makeStyles from '@mui/styles/makeStyles'
import { useTheme } from '@mui/material'
import { getParameterByName } from '../../util/query_param'
import { pathPrefix } from '../../env'

import logoSrcSet1 from '../../public/logos/black/goalert-logo-scaled.webp'
import logoSrcSet2 from '../../public/logos/black/goalert-logo-scaled@1.5.webp'
import logoSrcSet3 from '../../public/logos/black/goalert-logo-scaled@2.webp'
import logoImgSrc from '../../public/logos/black/goalert-logo-scaled@2.png'

import darkModeLogoSrcSet1 from '../../public/logos/white/goalert-logo-white-scaled.webp'
import darkModeLogoSrcSet2 from '../../public/logos/white/goalert-logo-white-scaled@1.5.webp'
import darkModeLogoSrcSet3 from '../../public/logos/white/goalert-logo-white-scaled@2.webp'
import darkModeLogoImgSrc from '../../public/logos/white/goalert-logo-white-scaled@2.png'

const PROVIDERS_URL = pathPrefix + '/api/v2/identity/providers'

const useStyles = makeStyles({
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
})
type Field = {
  ID: string
  Label: string
  Password: string
  Required: boolean
}
type Provider = {
  ID: string
  Fields: Field[]
  Hidden: boolean
  LogoUrl: string
  Title: string
  URL: string
}

export default function Login(): JSX.Element {
  const classes = useStyles()
  const theme = useTheme()
  const [error, setError] = useState(getParameterByName('login_error') || '')
  const [providers, setProviders] = useState([])

  useEffect(() => {
    // get providers
    fetch(PROVIDERS_URL)
      .then((res) => res.json())
      .then((data) => setProviders(data))
      .catch((err) => setError(err))
  }, [])

  /*
   * Renders a field from a provider
   */

  function renderField(field: Field): JSX.Element {
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
  function renderHasNextDivider(
    idx: number,
    len: number,
  ): JSX.Element | undefined {
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

  function renderProvider(
    provider: Provider,
    idx: number,
    len: number,
  ): JSX.Element | null {
    const {
      ID: id, // unique identifier of the provider
      Fields: fields, // holds a list of fields to include with the request
      Hidden: hidden, // indicates that the provider is not intended for user visibility
      LogoUrl: logoUrl, // optional URL of an icon to display with the provider
      Title: title, // user-viable string for identifying this provider
      URL: url, // the location of the form action (POST)
    } = provider
    if (hidden) return null

    // create login button
    let loginButton = null
    const loginIcon = logoUrl ? (
      <img alt='GoAlert' src={logoUrl} className={classes.loginIcon} />
    ) : null
    if (fields) {
      loginButton = (
        <Button type='submit' variant='contained'>
          Login
        </Button>
      )
    } else {
      loginButton = (
        <Button type='submit' variant='contained'>
          {loginIcon}
          Login with {title}
        </Button>
      )
    }

    let form = null
    if (fields && fields.length) {
      form = (
        <Grid container spacing={2}>
          {fields.map((field: Field) => renderField(field))}
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
        {renderHasNextDivider(idx, len)}
      </React.Fragment>
    )
  }

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

  const logo =
    theme.palette.mode === 'dark' ? (
      <picture>
        <source
          srcSet={`
            ${darkModeLogoSrcSet1},
            ${darkModeLogoSrcSet2} 1.5x,
            ${darkModeLogoSrcSet3} 2x
          `}
          type='image/webp'
        />
        <img src={darkModeLogoImgSrc} height={61} alt='GoAlert' />
      </picture>
    ) : (
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
    <div className={classes.center}>
      <Card className={classes.card}>
        <CardContent>
          <Grid container spacing={2} className={classes.gridContainer}>
            <Grid item xs={12}>
              {logo}
            </Grid>
            {providers.map((provider, idx) =>
              renderProvider(provider, idx, providers.length),
            )}
            {errorJSX}
          </Grid>
        </CardContent>
      </Card>
    </div>
  )
}
