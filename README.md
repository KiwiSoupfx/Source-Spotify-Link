# Source-Spotify-Link
Automatically updates a cfg file in your game folder (If you give it the right path). Depending on your setup, you may need to manually create the file.

## Env
Make sure you rename the included file to .env or make your own

Fill out your client id and secret from your app on https://developer.spotify.com/dashboard/. Make sure you use 'http://localhost:8080' for your redirect uri.

Don't have an app? Make one, it's easy :)

escaped_cfg_file_path="Fullpath\\game\\cfg\\listening.cfg"

max_errors are the max errors we will tolerate before force closing. Use -1 to disable.

custom_message NEEDS to include the command like say or echo. Make sure you use {SongName} and {Artists} if you want them in your message.

## Running the app

Navigate to http://localhost:8080 and the app should take you everywhere you need to go, provided you gave a valid client_id and secret.
Make sure to click the "Spotify Authorization" button and login if you're logged out.

Use a bind or type "exec config_name.cfg" into console to run the command.

More customization coming soon.