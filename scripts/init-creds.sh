#!/bin/bash

# Decode GOOGLE_CREDENTIALS_B64
if [[ -n "$GOOGLE_CREDENTIALS_B64" ]]; then
  mkdir -p /app/credentials
  echo "$GOOGLE_CREDENTIALS_B64" | base64 -d > /app/credentials/google-credentials.json
  export GOOGLE_APPLICATION_CREDENTIALS=/app/credentials/google-credentials.json
  echo "✅ Google credentials written to /app/credentials"
else
  echo "⚠️ GOOGLE_CREDENTIALS_B64 not set"
fi

# Start Go app
exec ./main
