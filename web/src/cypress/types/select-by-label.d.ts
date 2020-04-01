declare namespace Cypress {
  interface Chainable<Subject> {
    /**
     * Selects an item from a dropdown by it's label. Automatically accounts for search-selects.
     */
    selectByLabel: selectByLabelFn

    /**
     * Finds an item from a dropdown by it's label. Automatically accounts for search-selects.
     */
    findByLabel: findByLabelFn

    /**
     * Finds an item from a dropdown by it's label and removes it if it is a multiselect.
     */
    multiRemoveByLabel: multiRemoveByLabelFn
  }
}

type selectByLabelFn = (label: string) => Cypress.Chainable
type findByLabelFn = (label: string) => Cypress.Chainable
type multiRemoveByLabelFn = (label: string) => Cypress.Chainable
