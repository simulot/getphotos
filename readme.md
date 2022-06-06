This module aims to download photos from your smartphone into you photo library in folders based on photo's time.

It uses the `gio` command for accessing to attached devices and perform file operations. (see `man gio`)

# usage
```
Usage of getphotos:
  -device value
        device mount point. leave empty to search all devices mounted
  -exclude value
        exclude files containing the string. Can be repeated. (default .trashed,screenshoot)
  -library value
        photo libray path. (default /home/$user/Pictures)
  -move
        move files instead of copy them.
  -subfolder value
        Photo destination sub folder based on exposure date, accepts %Y,%m,%d for year month day, %H,%M,%S for hours minutes secondes (default Photos/%Y/%Y.%m/%Y.%m.%d)
```

# device option
By default, `getphotos` scans all attached devices having a `DCIM` folder.
It has been tested on a linux mint box with:
- iPhone 6s and +
- Android 12

The `-device` option indicates the path where photos are stored. When it's given, scan for attached device is skipped.


# move option
When the `-move` is given, the photos are removed from the device after a successful copy into the library.

# Photos library

The photo library is placed by default in your's Picture directory. The default location is /home/$user/Pictures. This can be changed with the option `-library`

Photos are placed in the library in a directory determined by the date of the photo. The default directory for a photo taken on 01/06/2022 is `/home/$user/Photos/2022/2022.06/2022.06.01`.

This can be changed using the `-subfolder` option. Following placeholder can be used:
- %Y for the year (ex. 2022)
- %m for the month (ex. 06)
- %d for the day (ex. 01)
- %H for the hour (ex. 16)
- %M for the minute (ex. 02)
- %S for de seconde (ex. 45)

