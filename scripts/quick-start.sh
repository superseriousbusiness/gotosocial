#!/bin/sh

set -e

if [ "$GTS_ADMIN_USERNAME" != "" ]
then
    /gotosocial/gotosocial admin account confirm --username ${GTS_ADMIN_USERNAME}
    ADMIN_NOT_EXISTS=$?

    if [ $ADMIN_NOT_EXISTS == 1 ]
    then
        echo "Initializing Admin user ${GTS_ADMIN_USERNAME}"
        /gotosocial/gotosocial admin account create --username ${GTS_ADMIN_USERNAME} --email ${GTS_ADMIN_EMAIL} --password '${GTS_ADMIN_PASSWORD}'

        /gotosocial/gotosocial admin account promote --username ${GTS_ADMIN_USERNAME}
    else
        echo "Existing admin user ${GTS_ADMIN_USERNAME} found"
    fi
fi

/gotosocial/gotosocial server start
