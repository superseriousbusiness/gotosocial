#!/bin/bash

docker build -t "superseriousbusiness/gotosocial:$(cat version)" .
