# Codemods

This directory has a collection of codemod scripts for doing cleanup/updates to the codebase.

## Running

```bash
# make sure you have jscodeshift installed
yarn global add jscodeshift

# use -t to reference the script, then give it the path(s) to directories or js files to update
#
# this example will apploy the relpath codemod to everything under ./app
jscodeshift -t ./codemods/relpath.js ./app
```
