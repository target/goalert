# GoAlert All-In-One Container

This directory contains the `Dockerfile` for building GoAlert's all-in-one (demo) Docker container.  
This container provides a simple way to start and explore GoAlert. It is not recommended for production use.

### Assumptions

`goalert` binary built with `GOOS=linux BUNDLE=1` located in this directory before docker build.   
`init.sql` PostgreSQL demo data located in this directory before docker build.
