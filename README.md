## Futurehome Sonos adapter
Unfinished

### Service name

`media_player`

### Interfaces

Type        | Interface                 | Value type        | Description
------------|---------------------------|-------------------|-------
in          | cmd.playback.set          | string            | play/pause/togglePlayPause/skipToNextTrack/skipToPreviousTrack
in          | cmd.playback.get_report   |                   |
out         | evt.playback.report       | string            |
-|||
in          | cmd.mode.set              | str_map           | {"repeat": false, "repeatOne": false, "crossfade": false, "shuffle": false}
in          | cmd.mode.get_report       |                   | 
out         | evt.mode.report           | str_map           |
-|||
in          | cmd.volume.set            | int               | 0-100
in          | cmd.volume.get_report     |                   |
out         | evt.volume.report         | str_map           | {«volume»: 85, «muted»: false}
-|||
in          | cmd.mute.set              | bool              |
out         | evt.mute.report           | bool              | true, false
-|||
in          | cmd.metadata.get_report   |                   | 
out         | evt.metadata.report       | str_map           | Album name, track name, imageUrl, artist etc. 

### Service props

Name           | Value example                                                      | Description
---------------|--------------------------------------------------------------------|-------
`sup_modes`    | repeat, repeatOne, shuffle, crossfade                              | supported modes. 
`sup_playback` | play, pause, togglePlayPause, skipTonNexTrack, skipToPreviousTrack | supported playbacks. 

More interfaces coming
