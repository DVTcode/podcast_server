#!/bin/bash

# Decode GOOGLE_CREDENTIALS_B64
if [[ -n "$GOOGLE_CREDENTIALS_B64" ]]; then
  echo "$GOOGLE_CREDENTIALS_B64" | base64 -d > /root/credentials/google-credentials.json
  export GOOGLE_APPLICATION_CREDENTIALS=/root/credentials/google-credentials.json
  echo "✅ Google credentials written"
else
  echo "⚠️ GOOGLE_CREDENTIALS_B64 not set"
fi

# Start Go app
exec ./main
