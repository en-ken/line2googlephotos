#!/bin/bash

ACCESS_TOKEN=$(curl --data "refresh_token=$OAUTH2_REFRESH_TOKEN" \
    --data "client_id=$OAUTH2_CLIENT_ID" \
    --data "client_secret=$OAUTH2_CLIENT_SECRET" \
    --data "grant_type=refresh_token" \
    https://www.googleapis.com/oauth2/v4/token | \
    jq '.access_token')

curl -H "Authorization: Bearer ${ACCESS_TOKEN}" \
    -H "Content-Type: application/json" \
    --data "{\"album\":{\"title\": \"${ALBUM_TITLE}\"}}" \
    https://photoslibrary.googleapis.com/v1/albums