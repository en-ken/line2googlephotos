#!/bin/bash
gcloud functions deploy line2googlephotos \
    --entry-point MessageReceived \
    --runtime go113 \
    --trigger-http \
    --allow-unauthenticated \
    --env-vars-file .env.yaml \
    --region ${REGION} \
    --project ${PROJECT}