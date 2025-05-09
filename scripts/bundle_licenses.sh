#!/bin/sh
set -ex

# Split on
# new line
IFS='
'

license_filter() { grep -iE '.*/license(\.\w+)?$'; }

OUTPUT='./web/assets/all_licenses.txt'
echo 'The GoToSocial software uses the following dependencies, whose licenses are reproduced in full:' > "$OUTPUT" # reset output
echo >> "$OUTPUT"

# Copy over any licenses in our golang dependencies
for file in $(find ./vendor | license_filter); do
    echo "----------------------------------------------------------" >> "$OUTPUT"
    echo                                                              >> "$OUTPUT"
    echo "${file}:"                                                   >> "$OUTPUT"
    echo                                                              >> "$OUTPUT"
    cat  "${file}"                                                    >> "$OUTPUT"
    echo                                                              >> "$OUTPUT"
done

# Copy over any licenses in our javacsript dependencies
for file in $(find ./web/source | license_filter); do
    echo "----------------------------------------------------------" >> "$OUTPUT"
    echo                                                              >> "$OUTPUT"
    echo "${file}:"                                                   >> "$OUTPUT"
    echo                                                              >> "$OUTPUT"
    cat  "${file}"                                                    >> "$OUTPUT"
    echo                                                              >> "$OUTPUT"
done

# Copy over misc other licenses
for file in ./LICENSE \
            ./web/source/nollamasworker/sha256.js; do
    echo "----------------------------------------------------------" >> "$OUTPUT"
    echo                                                              >> "$OUTPUT"
    echo "${file}:"                                                   >> "$OUTPUT"
    echo                                                              >> "$OUTPUT"
    cat  "${file}"                                                    >> "$OUTPUT"
    echo                                                              >> "$OUTPUT"
done
