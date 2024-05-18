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

If you haven't looked through my comments or spotify's api docs, you'll want to go to:
https://accounts.spotify.com/en/authorize?client_id=Your_ClientId_Here&redirect_uri=http%3A%2F%2Flocalhost%3A8080&response_type=code&scope=user-read-currently-playing Make sure you replace the 'Your_ClientId_Here' with your app's client ID.
After it loads, click the OK and it'll start getting track data and you can close the tab.

Use a bind or type "exec config_name.cfg" into console to run the command

More customization hopefully coming soon. Thinking about making a web ui.