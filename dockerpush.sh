#!/bin/bash

docker push "superseriousbusiness/gotosocial:$(git rev-parse --abbrev-ref HEAD)"
