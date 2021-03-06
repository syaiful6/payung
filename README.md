# Payung

Payung is a system utility backup for Linux, Mac OS X and BSD.

This project has beed forked from [gobackup](https://github.com/huacnlee/gobackup) since 2020.

## Current support status

### Databases

- MySQL
- PostgreSQL
- Redis
- MongoDB

### Compressor
- Gzip
- Brotli

### Storages
- Local
- FTP
- SCP
- Amazon S3
- Alibaba Cloud Object Storage Service (OSS)
- Dropbox

### Notification
- Slack

## Installation

You can download our current build in this repository's [release page](https://github.com/syaiful6/payung/releases) and place executable in your $PATH, ie: `/usr/local/bin/`.

## Configuration

Payung will seek config files in:
- ~/.payung/payung.yml
- /etc/payung/payung.yml

Example config file: [payung-reference.yml](https://github.com/syaiful6/payung/blob/develop/payung-reference.yml).

```yml
models:
  gitlab:
    compress_with:
      type: brotli
      level: 8
    store_with:
      type: s3
      bucket: backups
      region: us-east-1
      access_key_id: xxxxxx
      secret_access_key: xxxxxxx
      max_retries: 5
    notify_with:
      slack:
        type: slack
        webhook_url: https://hooks.slack.com/services/xxxxx
        channel: database_backups
        send_on:
          - success
          - failed
    databases:
      gitlab:
        type: mysql
        host: localhost
        port: 3306
        database: gitlab
        username: root
        password: xxxxxx
        additional_options: --single-transaction --quick --max_allowed_packet=1G
    archive:
      includes:
        - /home/git/.ssh/
        - /etc/mysql/my.conf
        - /etc/logrotate.d/
      excludes:
        - /home/ubuntu/.ssh/known_hosts
        - /etc/logrotate.d/syslog
```

## Usage

You can learn available options with `payung --help` and `payung perform --help`.

For example this is how you provides custom config file and dump path folder.

```
payung perform -c ~/backups/payung.yml -d ~/backups/workdir
```

We recommend you to perform backups using cron, here are an example configuration:

```
5 0 * * * /usr/local/bin/payung perform -c /home/ubuntu/payung.yml -d /mnt/ebs/backups >> /mnt/ebs/backups/backup.log 2>&1
```

### Recommendation

If you use Amazon S3, we recommends to use [AWS S3 object lifecycle](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html) instead of our storage lifecycle. You can use it to move the object to Glacier or delete the backups if older than one month.
