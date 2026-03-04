# Youtube Subscriptions RSS

Create RSS feeds of your Youtube subscriptions.

## Features

- Generate an xml feed of the latest videos for all your YouTube subscriptions
- Run a small web server for serving the feed
- One time generation of feed

## Configuration

### 1. Configure Google Cloud YouTube Data API

- Go to the [Google Cloud Console](https://console.cloud.google.com/apis/dashboard).
- Create a new project (or select an existing one).
- Enable the **YouTube Data API v3** for your project.
- Go to **APIs & Services → Credentials**.
- Click **Create Credentials → OAuth client ID**.
- Choose "Desktop app" and get your `client_id` and `client_secret`.
- Download the generated `credentials.json` file.

### 2. Place `credentials.json` in the Configs Folder

- Move your `credentials.json` file into a new folder called `configs` at the root of this repo. For example:

  ```
  configs/credentials.json
  ```

### 3. Create refresh token

Run this command and follow the steps to generate a refresh token

### 4. Place credentials in `.env` file

  - `YT_CLIENT_ID` — from your `credentials.json`
  - `YT_CLIENT_SECRET` — from your `credentials.json`
  - `YT_REFRESH_TOKEN` — from the refresh token command
    
You can optionally set this if you want to include shorts in the feed:

- `YT_RSS_INCLUDE_SHORTS=true`

## Commands

### One time feed generation

- `go run . generate-feed` or
- `mise generate-feed`

### Run server

- `go run . server` or
- `mise server`

### Generate refresh token

- `go run . refresh` or
- `mise refresh`

## Thanks

This was based primarily on this project: 
https://github.com/beviz/youtube-subscriptions-to-rss-feeds
