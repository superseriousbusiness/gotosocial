#!/bin/sh

# Mime types JSON source
URL='https://raw.githubusercontent.com/micnic/mime.json/master/index.json'

# Define intro to file
FILE='
// This is an automatically generated file, do not edit
package mimetypes


// MimeTypes is a map of file extensions to mime types.
var MimeTypes = map[string]string{
'

# Set break on new-line
IFS='
'

for line in $(curl -fL "$URL" | grep -E '".+"\s*:\s*".+"'); do
    # Trim final whitespace
    line=$(echo "$line" | sed -e 's|\s*$||')

    # Ensure it ends in a comma
    [ "${line%,}" = "$line" ] && line="${line},"

    # Add to file
    FILE="${FILE}${line}
"
done

# Add final statement to file
FILE="${FILE}
}

"

# Write to file
echo "$FILE" > 'mime.gen.go'

# Check for valid go
gofumpt -w 'mime.gen.go'