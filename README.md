# line2googlephotos
cloud function event implementation for uploading from LINE message to Google Photos

## Setup

0. Create an album and take a note of album ID.

```:bash
OAUTH2_REFRESH_TOKEN=[Refresh Token for Google Photos APIs call] ./create_album.sh
```

1. Set environment variables. (Create .env.yaml)

```:bash
ALBUM_ID: [Album ID of Google Photos]
LINE_CHANNEL_SECRET: [LINE Channel Secret]
LINE_CHANNEL_ACCESS_TOKEN: [LINE Channel Access Token]
OAUTH2_CLIENT_ID: [Client ID for Google OAuth2]
OAUTH2_CLIENT_SECRET: [Client Secret for Google OAuth2]
OAUTH2_REFRESH_TOKEN: [Refresh Token for Google Photos APIs call]
```


2. Deploy
```:bash
REGION=[GCP Region] PROJECT=[GCP Project Name] ./deploy.sh
```
