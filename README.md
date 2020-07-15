## Futurehome Sonos adapter
Unfinished

### Service name

`media_player`

### Interfaces

Type        | Interface                 | Value type        | Description
------------|---------------------------|-------------------|-------
in          | cmd.playback.set          | string            | play/pause/togglePlayPause/skipToNextTrack/skipToPreviousTrack
in          | cmd.playback_mode.set     | str_map           | {"repeat": false, "repeatOne": false, "crossfade": false, "shuffle": false}
in          | cmd.playback.get_report   |                   |
out         | evt.playback.report       | str_map           |
-|||
in          | cmd.volume.set            | int               | 0-10
in          | cmd.volume.get            |                   |
out         | evt.volume.report         | str_map           | {«volume»: 85, «muted»: false, «fixed»: false}
-|||

More interfaces coming
