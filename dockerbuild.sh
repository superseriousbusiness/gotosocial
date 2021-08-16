#!/bin/bash

docker build -t "superseriousbusiness/gotosocial:$(git rev-parse --abbrev-ref HEAD)" .
