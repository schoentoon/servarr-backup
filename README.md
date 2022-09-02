# Servarr-backup

This tool will help making backups of [radarr](https://github.com/Radarr/Radarr), [sonarr](https://github.com/Sonarr/Sonarr) and [lidarr](https://github.com/lidarr/Lidarr) into [restic](https://github.com/restic/restic), [borg](https://github.com/borgbackup/borg), whatever tool easier.

## Download

If you have golang installed you can install this by using `go install github.com/schoentoon/servarr-backup/cmd/servarr-backup@latest`.
Otherwise you can download it from the [Gitlab CI](https://gitlab.com/schoentoon/servarr-backup/-/pipelines/latest).

## Usage

```asciidoc
Usage of servarr-backup:
  -apikey string
    	Api key for the servarr
  -apiversion int
    	Set the api version, this should be 1 for lidarr and 3 for radarr/sonarr (default 3)
  -baseurl string
    	Base url of the servarr
  -delete
    	Should the backup be deleted from the servarr afterwards?
  -extract
    	Should we extract the zip file?
  -output string
    	Where to output the zip file to (- is stdout) (default "-")
```

## Usage example

Below is an example of how to use this tool together with restic.
You will obviously have to fill in the blanks like API KEY, BASE URL etc.
We are using a static temporary directory here, because restic groups snapshots by hostname and directory.
So if you were to extract into a true temporary directory, a `restic forget` wouldn't forget about any of the snapshots.
The reasoning behind extracting the zip files from sonarr, radarr, lidarr is to work nicely with the deduplication and compression.
And especially as the database files tend to only change tiny bits at the time, deduplication can work quite nicely, backing up the zip files directly however kind of ruins this.

```bash
tmp_dir="/tmp/sonarr"
mkdir "${tmp_dir}"

servarr-backup -apikey <API KEY> -apiversion 3 -baseurl <BASE URL> -extract -output "${tmp_dir}" -delete

restic -r <RESTIC REPOSITORY> backup "${tmp_dir}

rm -rf "${tmp_dir}"
```
