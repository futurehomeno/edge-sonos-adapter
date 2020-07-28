## Futurehome Sonos adapter
The adapter works with Futurehome, but is currentlyt not supported in the app. For now you can use Thingsplex to make flows and control your Sonos devices. 
Click [here](https://github.com/thingsplex/sonos-ad/blob/master/sonos_flow_example.json) to see example of a Sonos flow. 

### Service name
`media_player`
### Interfaces
Type        | Interface                 | Value type        | Description
------------|---------------------------|-------------------|-------
in          | cmd.playback.set          | string            | play, pause, toggle_play_pause, next_track, previous_track
in          | cmd.playback.get_report   | null              |
out         | evt.playback.report       | string            |
-|||
in          | cmd.mode.set              | str_map           | {"repeat": false, "repeat_one": false, "crossfade": false, "shuffle": false}
in          | cmd.mode.get_report       | null              | 
out         | evt.mode.report           | str_map           |
-|||
in          | cmd.volume.set            | int               | 0-100
in          | cmd.volume.get_report     | null              |
out         | evt.volume.report         | int               | 0-100
-|||
in          | cmd.mute.set              | bool              |
in          | cmd.mute.get_report       | null              |
out         | evt.mute.report           | bool              |
-|||
in          | cmd.metadata.get_report   | null              | 
out         | evt.metadata.report       | str_map           | {"album": "", "track": "", "artist": "", "image_url": ""}

### Service props
Name           | Value example                                                      | Description
---------------|--------------------------------------------------------------------|-------
`sup_modes`    | repeat, repeat_one, shuffle, crossfade                             | supported modes. 
`sup_playback` | play, pause, toggle_play_pause, next_track, previous_track         | supported playbacks.
`sup_metadata` | album, track, artist, image_url                                    | supported metadata. 
