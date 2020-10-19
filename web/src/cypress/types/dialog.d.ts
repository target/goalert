declare namespace Cypress {
  interface Chainable {
    /** Click a dialog button with the given text and wait for it to disappear. */
    dialogFinish: typeof dialogFinish

    /** Click a dialog button with the given text. */
    dialogClick: typeof dialogClick

    /** Assert a dialog is present with the given title string. */
    dialogTitle: typeof dialogTitle

    /** Assert a dialog with the given content is present. */
    dialogContains: typeof dialogContains

    /** Update a dialog's form fields with the given values. */
    dialogForm: typeof dialogForm

    /** Gets the dialog container. */
    dialog: typeof dialog
  }
}
