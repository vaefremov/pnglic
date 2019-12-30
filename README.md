# pnglic: Example of REST Backend and templates-based Frontend

Utility REST server to manage Pangea licenses and customers DB. Actually, this is an example
of REST server (API is  defined by the OpenAPI YAML descriptor) and a frontend
part using the API and built using the HTML templates.

## Overview

This is a backend server meeting the following requirements:

- Should use the current database
- CRUD Customers
- CRUD hardware license keys - both HASP and Gardian
- Manage distribution of keys over customers, including operation like transferring key from one customer to another
- Generate license files of current kind
- Generate license files supporting additional keys for end-users
- Support the history of issues for license files, including requests to search in the history

The aim of this work is the test the whole set of technologies of developing backend server using:
- OpenAPI service description
- Code generators from the OpenAPI
- Implementing server in Golang (HTTP/1.1 REST interface)
- Migrating to gRPC (HTTP/2)
- Implementing the original REST API using a reverse-proxy server

