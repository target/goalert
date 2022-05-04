# Cypress

Integration tests can be found under `./integration`. There are a variety of helper functions that can be used in your integration tests that can be found under `./support`. Mock data for the integration tests is under `./fixtures`.

All of the created support functions can be chained off of cypress, for example:

```
cy.get('input[name=someInput]').selectByLabel('Label Name')
```

A fixture can be used as such to access some mock data, for example:

```js
import users from '../fixtures/users.json'
```
